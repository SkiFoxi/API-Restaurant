package worker

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"restaurant-demo/internal/client"
)

type Worker struct {
	coreClient   *client.CoreClient
	restaurantID int64
	interval     time.Duration
	logger       *slog.Logger
	orderCh      chan int64
	workerCount  int

	processing map[int64]bool
	mu         sync.Mutex
	wg         sync.WaitGroup
	closeOnce  sync.Once

	// polling == 1, если предыдущий pollOrders ещё не завершился.
	// Нужно, чтобы медленный poll (например, из-за большого числа
	// заказов и ограниченной пропускной способности воркеров) не
	// блокировал основной цикл с тикером.
	polling int32
	// pollWg считает запущенные горутины pollOrdersAsync, чтобы
	// корректно дождаться их завершения при остановке.
	pollWg sync.WaitGroup
}

func NewWorker(
	coreClient *client.CoreClient,
	restaurantID int64,
	interval time.Duration,
	logger *slog.Logger,
	workerCount int,
) *Worker {
	if workerCount <= 0 {
		workerCount = 1
	}
	if interval <= 0 {
		interval = 5 * time.Second
	}
	return &Worker{
		coreClient:   coreClient,
		restaurantID: restaurantID,
		interval:     interval,
		logger:       logger,
		orderCh:      make(chan int64, 100),
		workerCount:  workerCount,
		processing:   make(map[int64]bool),
	}
}

// Run запускает worker и блокируется до отмены context.
func (w *Worker) Run(ctx context.Context) error {
	w.logger.Info(
		"Worker starting",
		"restaurant_id", w.restaurantID,
		"interval", w.interval,
		"workers", w.workerCount,
	)

	// Запускаем workers.
	for i := 0; i < w.workerCount; i++ {
		w.wg.Add(1)
		go w.worker(ctx, i)
	}
	w.logger.Info("Workers started", "count", w.workerCount)

	// Ticker вместо time.Sleep().
	// Теперь worker может немедленно отреагировать на ctx.Done().
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// После выхода из Run закрываем очередь и ждём workers.
	defer func() {
		w.logger.Info("Stopping workers")
		// Дожидаемся завершения всех запущенных горутин pollOrdersAsync,
		// чтобы не закрыть orderCh, пока в него ещё могут писать.
		w.pollWg.Wait()
		w.closeOnce.Do(func() { close(w.orderCh) })
		w.wg.Wait()
		w.logger.Info("All workers stopped")
	}()

	// Первый polling сразу после запуска.
	// Запускаем асинхронно (см. pollOrdersAsync) по той же причине,
	// что и в цикле ниже: pollOrders может надолго заблокироваться
	// на отправке в orderCh, если заказов пришло больше, чем воркеры
	// успевают обработать. Раньше это приводило к тому, что тикер
	// вообще не успевал начать тикать.
	w.pollOrdersAsync(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info(
				"Context cancelled, stopping main loop",
				"error", ctx.Err(),
			)
			return ctx.Err()
		case <-ticker.C:
			w.pollOrdersAsync(ctx)
		}
	}
}

// pollOrdersAsync запускает pollOrders в отдельной горутине, чтобы
// медленный poll (блокировка на переполненном orderCh) не мешал
// тикеру вовремя порождать следующие срабатывания.
//
// Если предыдущий poll ещё не завершился, новый тик пропускается —
// это ожидаемо: значит воркеры и так загружены под завязку, и смысла
// запускать ещё один параллельный опрос API нет.
func (w *Worker) pollOrdersAsync(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&w.polling, 0, 1) {
		w.logger.Warn("Previous poll is still running, skipping this tick")
		return
	}

	w.pollWg.Add(1)
	go func() {
		defer w.pollWg.Done()
		defer atomic.StoreInt32(&w.polling, 0)
		w.pollOrders(ctx)
	}()
}

// pollOrders получает новые заказы и отправляет их workers.
func (w *Worker) pollOrders(ctx context.Context) {
	// ДИАГНОСТИКА: запись в файл, чтобы проверить, вызывается ли функция
	f, _ := os.OpenFile("/tmp/poll.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	f.WriteString("polled at " + time.Now().String() + "\n")
	f.Close()
	// Если context уже отменён — не делаем запрос к API.
	select {
	case <-ctx.Done():
		return
	default:
	}

	w.logger.Info("Polling Core API for new orders")
	orders, err := w.coreClient.FetchNewOrders(w.restaurantID)
	if err != nil {
		w.logger.Error(
			"Error fetching orders",
			"error", err,
		)
		return
	}
	w.logger.Info(
		"Fetched orders",
		"count", len(orders),
	)

	for _, order := range orders {
		// Проверяем, не обрабатывается ли заказ уже.
		w.mu.Lock()
		if w.processing[order.ID] {
			w.mu.Unlock()
			w.logger.Debug(
				"Order already processing, skipping",
				"order_id", order.ID,
			)
			continue
		}
		// Сразу помечаем заказ как обрабатываемый.
		w.processing[order.ID] = true
		w.mu.Unlock()

		// Пытаемся поставить заказ в очередь.
		select {
		case w.orderCh <- order.ID:
			w.logger.Info(
				"Order queued for processing",
				"order_id", order.ID,
			)
		case <-ctx.Done():
			// Не смогли поставить заказ в очередь.
			// Убираем его из processing, чтобы его можно было
			// обработать повторно после следующего запуска.
			w.mu.Lock()
			delete(w.processing, order.ID)
			w.mu.Unlock()
			w.logger.Warn(
				"Context cancelled while enqueuing order",
				"order_id", order.ID,
			)
			return
		}
	}
}

// worker обрабатывает заказы из очереди.
func (w *Worker) worker(ctx context.Context, id int) {
	defer func() {
		if r := recover(); r != nil {
			w.logger.Error(
				"Worker panicked",
				"worker_id", id,
				"recover", r,
			)
		}
		w.wg.Done()
	}()

	w.logger.Info(
		"Worker goroutine started",
		"worker_id", id,
	)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info(
				"Worker context cancelled",
				"worker_id", id,
			)
			return
		case orderID, ok := <-w.orderCh:
			if !ok {
				w.logger.Info(
					"Order channel closed",
					"worker_id", id,
				)
				return
			}
			w.logger.Debug(
				"Worker received order",
				"worker_id", id,
				"order_id", orderID,
			)
			w.processOrder(ctx, orderID)
		}
	}
}

// processOrder обрабатывает один заказ.
func (w *Worker) processOrder(ctx context.Context, orderID int64) {
	// В любом случае удаляем заказ из processing.
	defer func() {
		w.mu.Lock()
		delete(w.processing, orderID)
		w.mu.Unlock()
	}()

	// Проверяем context перед началом обработки.
	select {
	case <-ctx.Done():
		w.logger.Warn(
			"Order processing cancelled before start",
			"order_id", orderID,
		)
		return
	default:
	}

	w.logger.Info(
		"Processing order",
		"order_id", orderID,
	)

	// 1. Переводим заказ в collected (принят к приготовлению)
	if err := w.coreClient.UpdateOrderStatus(orderID, "collected"); err != nil {
		w.logger.Error(
			"Failed to set collected",
			"order_id", orderID,
			"error", err,
		)
		return
	}
	w.logger.Info(
		"Order now collected",
		"order_id", orderID,
	)

	// Пауза 5 секунд (имитация подготовки)
	select {
	case <-time.After(5 * time.Second):
		// подготовка завершена
	case <-ctx.Done():
		w.logger.Warn(
			"Order processing cancelled while preparing",
			"order_id", orderID,
		)
		return
	}

	// 2. Переводим заказ в cooking
	if err := w.coreClient.UpdateOrderStatus(orderID, "cooking"); err != nil {
		w.logger.Error(
			"Failed to start cooking",
			"order_id", orderID,
			"error", err,
		)
		return
	}
	w.logger.Info(
		"Order now cooking",
		"order_id", orderID,
	)

	// Симулируем приготовление 10 секунд.
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	select {
	case <-timer.C:
		// Приготовление завершено.
	case <-ctx.Done():
		w.logger.Warn(
			"Order processing cancelled during cooking",
			"order_id", orderID,
		)
		return
	}

	// 3. Переводим заказ в ready
	if err := w.coreClient.UpdateOrderStatus(orderID, "ready"); err != nil {
		w.logger.Error(
			"Failed to mark order as ready",
			"order_id", orderID,
			"error", err,
		)
		return
	}
	w.logger.Info(
		"Order is ready",
		"order_id", orderID,
	)
}
