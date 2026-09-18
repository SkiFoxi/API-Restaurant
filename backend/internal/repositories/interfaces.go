// Интерфейсы для репозиториев
// Помогают тестировать приложение и удобны для передачи в бизнес-логику
package repositories

import (
	"context"

	"avito-kitchen/backend/internal/models"
)

// Интерфейс для работы с ресторанами
type RestaurantRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Restaurant, error)
	GetByName(ctx context.Context, name string) (*models.Restaurant, error)
	Create(ctx context.Context, restaurant *models.Restaurant) error
	List(ctx context.Context, limit int, cursor int64) ([]*models.Restaurant, int64, error)
}

// Интерфейс работы с меню ресторана
type MenuItemRepository interface {
	GetByID(ctx context.Context, id int64) (*models.MenuItem, error)
	ListByRestaurant(ctx context.Context, restaurantID int64, availableOnly bool) ([]*models.MenuItem, error)
	Create(ctx context.Context, item *models.MenuItem) error
	UpdateAvailability(ctx context.Context, id int64, available bool) error
}

// Интерфейс работы с заказами
type OrderRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Order, error)
	Create(ctx context.Context, order *models.Order, items []*models.OrderItem) error // транзакция
	ListByRestaurant(ctx context.Context, restaurantID int64, status string, cursor int64, limit int) ([]*models.Order, int64, error)
	UpdateStatus(ctx context.Context, orderID int64, status string) error
	Update(ctx context.Context, order *models.Order) error
}
