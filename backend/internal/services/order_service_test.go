package services

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"avito-kitchen/backend/internal/models"
	"avito-kitchen/backend/internal/repositories"
)

// --- Моки ---

type mockOrderRepo struct {
	createFunc       func(ctx context.Context, order *models.Order, items []*models.OrderItem) error
	getByIDFunc      func(ctx context.Context, id int64) (*models.Order, error)
	listByRestFunc   func(ctx context.Context, restaurantID int64, status string, cursor int64, limit int) ([]*models.Order, int64, error)
	updateStatusFunc func(ctx context.Context, orderID int64, status string) error
	updateFunc       func(ctx context.Context, order *models.Order) error
}

func (m *mockOrderRepo) Create(ctx context.Context, order *models.Order, items []*models.OrderItem) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, order, items)
	}
	return nil
}
func (m *mockOrderRepo) GetByID(ctx context.Context, id int64) (*models.Order, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, repositories.ErrNotFound
}
func (m *mockOrderRepo) ListByRestaurant(ctx context.Context, restaurantID int64, status string, cursor int64, limit int) ([]*models.Order, int64, error) {
	if m.listByRestFunc != nil {
		return m.listByRestFunc(ctx, restaurantID, status, cursor, limit)
	}
	return nil, 0, nil
}
func (m *mockOrderRepo) UpdateStatus(ctx context.Context, orderID int64, status string) error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, orderID, status)
	}
	return nil
}
func (m *mockOrderRepo) Update(ctx context.Context, order *models.Order) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, order)
	}
	return nil
}

type mockMenuRepo struct {
	getByIDFunc      func(ctx context.Context, id int64) (*models.MenuItem, error)
	listByRestFunc   func(ctx context.Context, restaurantID int64, availableOnly bool) ([]*models.MenuItem, error)
	createFunc       func(ctx context.Context, item *models.MenuItem) error
	updateAvailFunc  func(ctx context.Context, id int64, available bool) error
}

func (m *mockMenuRepo) GetByID(ctx context.Context, id int64) (*models.MenuItem, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, repositories.ErrNotFound
}
func (m *mockMenuRepo) ListByRestaurant(ctx context.Context, restaurantID int64, availableOnly bool) ([]*models.MenuItem, error) {
	if m.listByRestFunc != nil {
		return m.listByRestFunc(ctx, restaurantID, availableOnly)
	}
	return nil, nil
}
func (m *mockMenuRepo) Create(ctx context.Context, item *models.MenuItem) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, item)
	}
	return nil
}
func (m *mockMenuRepo) UpdateAvailability(ctx context.Context, id int64, available bool) error {
	if m.updateAvailFunc != nil {
		return m.updateAvailFunc(ctx, id, available)
	}
	return nil
}

// Исправленный мок: добавлен метод GetByName
type mockRestRepo struct {
	getByIDFunc   func(ctx context.Context, id int64) (*models.Restaurant, error)
	getByNameFunc func(ctx context.Context, name string) (*models.Restaurant, error)
	createFunc    func(ctx context.Context, restaurant *models.Restaurant) error
	listFunc      func(ctx context.Context, limit int, cursor int64) ([]*models.Restaurant, int64, error)
}

func (m *mockRestRepo) GetByID(ctx context.Context, id int64) (*models.Restaurant, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, repositories.ErrNotFound
}
func (m *mockRestRepo) GetByName(ctx context.Context, name string) (*models.Restaurant, error) {
	if m.getByNameFunc != nil {
		return m.getByNameFunc(ctx, name)
	}
	return nil, repositories.ErrNotFound
}
func (m *mockRestRepo) Create(ctx context.Context, restaurant *models.Restaurant) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, restaurant)
	}
	return nil
}
func (m *mockRestRepo) List(ctx context.Context, limit int, cursor int64) ([]*models.Restaurant, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, limit, cursor)
	}
	return nil, 0, nil
}

// --- Вспомогательные функции ---

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// --- Тесты ---

func TestOrderService_CreateOrder_Success(t *testing.T) {
	restRepo := &mockRestRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Restaurant, error) {
			return &models.Restaurant{
				ID:           1,
				Name:         "Test Restaurant",
				IsActive:     true,
				OpeningTime:  "00:00:00",
				ClosingTime:  "23:59:59",
			}, nil
		},
	}
	menuRepo := &mockMenuRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.MenuItem, error) {
			return &models.MenuItem{
				ID:           id,
				RestaurantID: 1,
				Name:         "Pizza",
				Price:        100,
				IsAvailable:  true,
			}, nil
		},
	}
	var capturedItems []*models.OrderItem
	orderRepo := &mockOrderRepo{
		createFunc: func(ctx context.Context, order *models.Order, items []*models.OrderItem) error {
			capturedItems = items
			order.ID = 123
			return nil
		},
	}
	logger := testLogger()
	svc := NewOrderService(orderRepo, menuRepo, restRepo, logger)

	cart := []struct {
		MenuItemID int64
		Quantity   int
	}{
		{MenuItemID: 10, Quantity: 2},
	}
	order, err := svc.CreateOrder(context.Background(), 1, 1, "Test Address", cart)

	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, int64(123), order.ID)
	assert.Equal(t, float64(200), order.TotalPrice)
	assert.Equal(t, models.OrderStatusNew, order.Status)
	assert.Len(t, capturedItems, 1)
	assert.Equal(t, int64(10), capturedItems[0].MenuItemID)
	assert.Equal(t, 2, capturedItems[0].Quantity)
	assert.Equal(t, float64(100), capturedItems[0].PriceAtTime)
}

func TestOrderService_CreateOrder_RestaurantNotFound(t *testing.T) {
	restRepo := &mockRestRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Restaurant, error) {
			return nil, repositories.ErrNotFound
		},
	}
	menuRepo := &mockMenuRepo{}
	orderRepo := &mockOrderRepo{}
	logger := testLogger()
	svc := NewOrderService(orderRepo, menuRepo, restRepo, logger)

	cart := []struct {
		MenuItemID int64
		Quantity   int
	}{{1, 1}}
	order, err := svc.CreateOrder(context.Background(), 1, 1, "Addr", cart)

	assert.Nil(t, order)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "restaurant not found")
}

func TestOrderService_CreateOrder_MenuItemNotFound(t *testing.T) {
	restRepo := &mockRestRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Restaurant, error) {
			return &models.Restaurant{
				ID:           1,
				IsActive:     true,
				OpeningTime:  "00:00:00",
				ClosingTime:  "23:59:59",
			}, nil
		},
	}
	menuRepo := &mockMenuRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.MenuItem, error) {
			return nil, repositories.ErrNotFound
		},
	}
	orderRepo := &mockOrderRepo{}
	logger := testLogger()
	svc := NewOrderService(orderRepo, menuRepo, restRepo, logger)

	cart := []struct {
		MenuItemID int64
		Quantity   int
	}{{999, 1}}
	order, err := svc.CreateOrder(context.Background(), 1, 1, "Addr", cart)

	assert.Nil(t, order)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "menu item 999 not found")
}

func TestOrderService_CreateOrder_RestaurantClosed(t *testing.T) {
	// Используем относительное время, чтобы тест не был flaky
	now := time.Now().UTC()
	opening := now.Add(2 * time.Hour).Format("15:04:05")
	closing := now.Add(3 * time.Hour).Format("15:04:05")

	restRepo := &mockRestRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Restaurant, error) {
			return &models.Restaurant{
				ID:           1,
				IsActive:     true,
				OpeningTime:  opening,
				ClosingTime:  closing,
			}, nil
		},
	}
	menuRepo := &mockMenuRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.MenuItem, error) {
			return &models.MenuItem{ID: id, RestaurantID: 1, IsAvailable: true, Price: 100}, nil
		},
	}
	orderRepo := &mockOrderRepo{}
	logger := testLogger()
	svc := NewOrderService(orderRepo, menuRepo, restRepo, logger)

	cart := []struct {
		MenuItemID int64
		Quantity   int
	}{{1, 1}}
	order, err := svc.CreateOrder(context.Background(), 1, 1, "Addr", cart)

	assert.Nil(t, order)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "restaurant is closed at this time")
}

func TestOrderService_CreateOrder_NotAvailable(t *testing.T) {
	restRepo := &mockRestRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Restaurant, error) {
			return &models.Restaurant{
				ID:           1,
				IsActive:     true,
				OpeningTime:  "00:00:00",
				ClosingTime:  "23:59:59",
			}, nil
		},
	}
	menuRepo := &mockMenuRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.MenuItem, error) {
			return &models.MenuItem{ID: id, RestaurantID: 1, IsAvailable: false, Price: 100}, nil
		},
	}
	orderRepo := &mockOrderRepo{}
	logger := testLogger()
	svc := NewOrderService(orderRepo, menuRepo, restRepo, logger)

	cart := []struct {
		MenuItemID int64
		Quantity   int
	}{{1, 1}}
	order, err := svc.CreateOrder(context.Background(), 1, 1, "Addr", cart)

	assert.Nil(t, order)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "menu item 1 is not available")
}

func TestOrderService_GetOrderByID_Success(t *testing.T) {
	expected := &models.Order{ID: 1, Status: models.OrderStatusNew}
	orderRepo := &mockOrderRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Order, error) {
			if id == 1 {
				return expected, nil
			}
			return nil, repositories.ErrNotFound
		},
	}
	menuRepo := &mockMenuRepo{}
	restRepo := &mockRestRepo{}
	logger := testLogger()
	svc := NewOrderService(orderRepo, menuRepo, restRepo, logger)

	order, err := svc.GetOrderByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, expected, order)
}

func TestOrderService_GetOrderByID_NotFound(t *testing.T) {
	orderRepo := &mockOrderRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Order, error) {
			return nil, repositories.ErrNotFound
		},
	}
	menuRepo := &mockMenuRepo{}
	restRepo := &mockRestRepo{}
	logger := testLogger()
	svc := NewOrderService(orderRepo, menuRepo, restRepo, logger)

	order, err := svc.GetOrderByID(context.Background(), 99)

	assert.Nil(t, order)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, repositories.ErrNotFound))
}

func TestOrderService_UpdateOrderStatus_Valid(t *testing.T) {
	orderRepo := &mockOrderRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Order, error) {
			return &models.Order{ID: id, Status: models.OrderStatusNew}, nil
		},
		updateStatusFunc: func(ctx context.Context, orderID int64, status string) error {
			assert.Equal(t, int64(1), orderID)
			assert.Equal(t, models.OrderStatusCollected, status)
			return nil
		},
	}
	menuRepo := &mockMenuRepo{}
	restRepo := &mockRestRepo{}
	logger := testLogger()
	svc := NewOrderService(orderRepo, menuRepo, restRepo, logger)

	err := svc.UpdateOrderStatus(context.Background(), 1, models.OrderStatusCollected)
	assert.NoError(t, err)
}

func TestOrderService_UpdateOrderStatus_InvalidTransition(t *testing.T) {
	orderRepo := &mockOrderRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Order, error) {
			return &models.Order{ID: id, Status: models.OrderStatusDelivered}, nil
		},
	}
	menuRepo := &mockMenuRepo{}
	restRepo := &mockRestRepo{}
	logger := testLogger()
	svc := NewOrderService(orderRepo, menuRepo, restRepo, logger)

	err := svc.UpdateOrderStatus(context.Background(), 1, models.OrderStatusCooking)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status transition")
}