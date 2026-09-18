package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"restaurant-demo/internal/client"
	"restaurant-demo/internal/logger"
	"restaurant-demo/internal/models"
	"restaurant-demo/internal/worker"
)

func main() {
	// Настройка логгера (без изменений)
	log := logger.New(logger.Config{
		Level:      slog.LevelInfo,
		JSON:       false,
		AddContext: false,
		File: &logger.FileConfig{
			Path:       "demo.log",
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
		},
	})
	slog.SetDefault(log)

	coreURL := os.Getenv("CORE_API_URL")
	if coreURL == "" {
		coreURL = "http://backend:8080"
	}

	coreClient := client.NewCoreClient(coreURL, log)

	// Ищем уже зарегистрированный ресторан по имени — чтобы при каждом
	// рестарте контейнера не плодить новые записи (и, как следствие,
	// новый restaurant_id, для которого воркер начинает опрашивать
	// заказы "с чистого листа").
	rest := models.RestaurantRegistration{
		Name:        "Demo Restaurant",
		Address:     "ул. Тестовая, 1",
		Phone:       "+7(999)111-22-33",
		OpeningTime: "09:00:00",
		ClosingTime: "22:00:00",
	}

	restID, err := coreClient.FindRestaurantByName(rest.Name)
	if err != nil {
		log.Error("Failed to look up restaurant", "error", err)
		os.Exit(1)
	}

	if restID != 0 {
		log.Info("Reusing existing restaurant", "id", restID)
	} else {
		restID, err = coreClient.RegisterRestaurant(rest)
		if err != nil {
			log.Error("Failed to register restaurant", "error", err)
			os.Exit(1)
		}
		log.Info("Registered restaurant", "id", restID)

		// Меню создаём только для только что зарегистрированного ресторана —
		// если ресторан уже существовал, меню тоже наверняка уже создано.
		menuItems := []models.MenuItem{
			{RestaurantID: restID, Name: "Пицца Маргарита", Description: "Томатный соус, сыр", Price: 550},
			{RestaurantID: restID, Name: "Роллы Филадельфия", Description: "Лосось, авокадо, сыр", Price: 650},
			{RestaurantID: restID, Name: "Салат Цезарь", Description: "Курица, сыр, сухарики", Price: 450},
		}
		for _, item := range menuItems {
			if err := coreClient.CreateMenuItem(item); err != nil {
				log.Error("Failed to create menu item", "name", item.Name, "error", err)
			} else {
				log.Info("Created menu item", "name", item.Name)
			}
		}
	}

	// Создание воркера
	w := worker.NewWorker(coreClient, restID, 10*time.Second, log, 5)

	// Контекст для отмены
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Канал, который закроется, когда воркер завершится
	done := make(chan struct{})

	// Запускаем воркер в отдельной горутине
	go func() {
		log.Info("Starting order polling...")
		if err := w.Run(ctx); err != nil && err != context.Canceled {
			log.Error("Worker stopped with error", "error", err)
		}
		log.Info("Worker finished")
		close(done)
	}()

	// Ожидаем сигнал завершения
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Info("Received signal, shutting down", "signal", sig)
		cancel() // отменяем контекст
		<-done   // дожидаемся реального завершения воркера
		log.Info("Application stopped")
	case <-done:
		// Если воркер завершился сам (например, из-за фатальной ошибки), выходим
		log.Warn("Worker stopped unexpectedly")
		cancel()
	}
}