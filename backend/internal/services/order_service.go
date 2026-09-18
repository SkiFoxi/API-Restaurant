package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"avito-kitchen/backend/internal/models"
	"avito-kitchen/backend/internal/repositories"
)

type OrderService struct {
	orderRepo repositories.OrderRepository      // Работа с заказами
	menuRepo  repositories.MenuItemRepository   // Работа с меню
	restRepo  repositories.RestaurantRepository // Работа с ресторанами
	logger    *slog.Logger                      // Логгер
}

func NewOrderService(
	orderRepo repositories.OrderRepository,
	menuRepo repositories.MenuItemRepository,
	restRepo repositories.RestaurantRepository,
	logger *slog.Logger,
) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		menuRepo:  menuRepo,
		restRepo:  restRepo,
		logger:    logger,
	}
}

// CreateOrder – создание заказа (транзакция через репозиторий)
func (s *OrderService) CreateOrder(ctx context.Context, restaurantID, userID int64, address string, cartItems []struct {
	MenuItemID int64
	Quantity   int
}) (*models.Order, error) {
	s.logger.InfoContext(ctx, "Creating order",
		slog.Int64("restaurant_id", restaurantID),
		slog.Int64("user_id", userID),
		slog.Int("items_count", len(cartItems)),
	)

	// 1. Проверяем, что ресторан существует и активен
	rest, err := s.restRepo.GetByID(ctx, restaurantID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, fmt.Errorf("restaurant not found")
		}
		return nil, fmt.Errorf("failed to get restaurant: %w", err)
	}
	if !rest.IsActive {
		return nil, errors.New("restaurant is not active")
	}

	// 2. Проверяем, что ресторан работает в текущее время с учетом работы ночью
	if err := s.isRestaurantOpen(rest); err != nil {
		return nil, err
	}

	// 3. Собираем позиции и проверяем наличие
	var orderItems []models.OrderItem
	var totalPrice float64

	for _, cart := range cartItems {
		menuItem, err := s.menuRepo.GetByID(ctx, cart.MenuItemID)
		if err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return nil, fmt.Errorf("menu item %d not found", cart.MenuItemID)
			}
			return nil, fmt.Errorf("failed to get menu item: %w", err)
		}
		if !menuItem.IsAvailable {
			return nil, fmt.Errorf("menu item %d is not available", cart.MenuItemID)
		}
		if menuItem.RestaurantID != restaurantID {
			return nil, fmt.Errorf("menu item %d does not belong to this restaurant", cart.MenuItemID)
		}

		orderItems = append(orderItems, models.OrderItem{
			MenuItemID:  cart.MenuItemID,
			Quantity:    cart.Quantity,
			PriceAtTime: menuItem.Price,
		})
		totalPrice += menuItem.Price * float64(cart.Quantity)
	}

	// 4. Создаём заказ
	order := &models.Order{
		RestaurantID:        restaurantID,
		UserID:              userID,
		Address:             address,
		TotalPrice:          totalPrice,
		Status:              models.OrderStatusNew,
		EstimatedDeliveryAt: nil, // можно вычислить позже
	}

	// Преобразуем []OrderItem в []*OrderItem для репозитория
	itemsPtr := make([]*models.OrderItem, len(orderItems))
	for i := range orderItems {
		itemsPtr[i] = &orderItems[i]
	}

	if err := s.orderRepo.Create(ctx, order, itemsPtr); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	s.logger.InfoContext(ctx, "Order created successfully",
		//order.ID это первичный ключ в БД (генерируется автоматически)
		slog.Int64("order_id", order.ID),
		slog.Float64("total_price", order.TotalPrice),
	)
	return order, nil
}

// isRestaurantOpen – проверяет, работает ли ресторан в текущее время
// Поддерживает переход через полночь (например, 22:00–02:00)
func (s *OrderService) isRestaurantOpen(rest *models.Restaurant) error {
	now := time.Now()
	loc := now.Location()

	//Это не хардкод, а проверка по шаблону HH:MM:SS
	open, _ := time.ParseInLocation("15:04:05", rest.OpeningTime, loc)
	close, _ := time.ParseInLocation("15:04:05", rest.ClosingTime, loc)

	open = time.Date(now.Year(), now.Month(), now.Day(), open.Hour(), open.Minute(), open.Second(), 0, loc)
	close = time.Date(now.Year(), now.Month(), now.Day(), close.Hour(), close.Minute(), close.Second(), 0, loc)

	if close.Before(open) {
		close = close.Add(24 * time.Hour) // переход через полночь
	}

	if now.Before(open) || now.After(close) {
		return errors.New("restaurant is closed at this time")
	}
	return nil
}

// GetOrderByID – получение заказа с позициями (но позиции отдельно не грузим, только заказ)
func (s *OrderService) GetOrderByID(ctx context.Context, id int64) (*models.Order, error) {
	return s.orderRepo.GetByID(ctx, id)
}

// ListOrdersForRestaurant – список заказов для ресторана
func (s *OrderService) ListOrdersForRestaurant(ctx context.Context, restaurantID int64, status string, cursor int64, limit int) ([]*models.Order, int64, error) {
	s.logger.DebugContext(ctx, "Listing orders for restaurant",
		slog.Int64("restaurant_id", restaurantID),
		slog.String("status", status),
	)
	return s.orderRepo.ListByRestaurant(ctx, restaurantID, status, cursor, limit)
}

// UpdateOrderStatus – обновить статус (с проверкой допустимости перехода)
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID int64, newStatus string) error {
	s.logger.InfoContext(ctx, "Updating order status",
		slog.Int64("order_id", orderID),
		slog.String("new_status", newStatus),
	)

	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if !order.CanTransitionTo(newStatus) {
		return fmt.Errorf("invalid status transition from %s to %s", order.Status, newStatus)
	}

	return s.orderRepo.UpdateStatus(ctx, orderID, newStatus)
}

// CancelOrder – отмена заказа (если статус позволяет)
func (s *OrderService) CancelOrder(ctx context.Context, orderID int64) error {
	return s.UpdateOrderStatus(ctx, orderID, models.OrderStatusCancelled)
}
