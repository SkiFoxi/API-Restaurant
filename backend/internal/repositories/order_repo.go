// Реализация интерфейса OrderRepository
package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"avito-kitchen/backend/internal/models"
)

type orderRepo struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewOrderRepository(pool *pgxpool.Pool, logger *slog.Logger) OrderRepository {
	return &orderRepo{pool: pool, logger: logger}
}

// --- GetByID ---
func (r *orderRepo) GetByID(ctx context.Context, id int64) (*models.Order, error) {
	r.logger.DebugContext(ctx, "Fetching order by ID", slog.Int64("order_id", id))

	var order models.Order
	query := `SELECT id, restaurant_id, user_id, address, total_price, status, 
		estimated_delivery_at, created_at, updated_at FROM orders WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&order.ID, &order.RestaurantID, &order.UserID, &order.Address,
		&order.TotalPrice, &order.Status, &order.EstimatedDeliveryAt,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.WarnContext(ctx, "Order not found", slog.Int64("order_id", id))
			return nil, ErrNotFound
		}
		r.logger.ErrorContext(ctx, "Failed to get order", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get order by id: %w", err)
	}
	return &order, nil
}

// --- Create (в транзакции) ---
func (r *orderRepo) Create(ctx context.Context, order *models.Order, items []*models.OrderItem) error {
	r.logger.DebugContext(ctx, "Creating order with items",
		slog.Int64("restaurant_id", order.RestaurantID),
		slog.Int("items_count", len(items)),
	)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.logger.ErrorContext(ctx, "Failed to begin transaction", slog.String("error", err.Error()))
		return fmt.Errorf("begin tx: %w", err)
	}
	//Для проверки на ошибку
	defer func() {
    if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
        r.logger.Error("Rollback failed", "error", err)
    }
}()

	// Вставка заказа
	queryOrder := `INSERT INTO orders 
		(restaurant_id, user_id, address, total_price, status, estimated_delivery_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`
	err = tx.QueryRow(ctx, queryOrder,
		order.RestaurantID, order.UserID, order.Address,
		order.TotalPrice, order.Status, order.EstimatedDeliveryAt,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		r.logger.ErrorContext(ctx, "Failed to insert order", slog.String("error", err.Error()))
		return fmt.Errorf("insert order: %w", err)
	}

	// Вставка позиций заказа
	for _, item := range items {
		item.OrderID = order.ID
		queryItem := `INSERT INTO order_items (order_id, menu_item_id, quantity, price_at_time)
			VALUES ($1, $2, $3, $4)
			RETURNING id`
		err = tx.QueryRow(ctx, queryItem,
			item.OrderID, item.MenuItemID, item.Quantity, item.PriceAtTime,
		).Scan(&item.ID)
		if err != nil {
			r.logger.ErrorContext(ctx, "Failed to insert order item",
				slog.Int64("menu_item_id", item.MenuItemID),
				slog.String("error", err.Error()),
			)
			return fmt.Errorf("insert order item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		r.logger.ErrorContext(ctx, "Failed to commit transaction", slog.String("error", err.Error()))
		return fmt.Errorf("commit tx: %w", err)
	}

	r.logger.InfoContext(ctx, "Order created successfully",
		slog.Int64("order_id", order.ID),
		slog.Float64("total_price", order.TotalPrice),
	)
	return nil
}

// --- ListByRestaurant (пагинация по cursor) ---
func (r *orderRepo) ListByRestaurant(ctx context.Context, restaurantID int64, status string, cursor int64, limit int) ([]*models.Order, int64, error) {
	r.logger.DebugContext(ctx, "Listing orders for restaurant",
		slog.Int64("restaurant_id", restaurantID),
		slog.String("status", status),
		slog.Int64("cursor", cursor),
		slog.Int("limit", limit),
	)

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `SELECT id, restaurant_id, user_id, address, total_price, status,
		estimated_delivery_at, created_at, updated_at
		FROM orders
		WHERE restaurant_id = $1 AND id > $2`
	args := []any{restaurantID, cursor}

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", len(args)+1)
		args = append(args, status)
	}
	query += fmt.Sprintf(" ORDER BY id ASC LIMIT $%d", len(args)+1)
	args = append(args, limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		r.logger.ErrorContext(ctx, "Failed to list orders", slog.String("error", err.Error()))
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	var lastID int64
	for rows.Next() {
		var o models.Order
		err := rows.Scan(&o.ID, &o.RestaurantID, &o.UserID, &o.Address,
			&o.TotalPrice, &o.Status, &o.EstimatedDeliveryAt,
			&o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			r.logger.ErrorContext(ctx, "Failed to scan order row", slog.String("error", err.Error()))
			return nil, 0, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, &o)
		lastID = o.ID
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration: %w", err)
	}

	var nextCursor int64
	if len(orders) == limit {
		nextCursor = lastID
	} else {
		nextCursor = 0
	}
	return orders, nextCursor, nil
}

// --- UpdateStatus ---
func (r *orderRepo) UpdateStatus(ctx context.Context, orderID int64, status string) error {
	r.logger.DebugContext(ctx, "Updating order status",
		slog.Int64("order_id", orderID),
		slog.String("new_status", status),
	)

	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
	cmdTag, err := r.pool.Exec(ctx, query, status, orderID)
	if err != nil {
		r.logger.ErrorContext(ctx, "Failed to update order status", slog.String("error", err.Error()))
		return fmt.Errorf("update order status: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		r.logger.WarnContext(ctx, "Order not found for status update", slog.Int64("order_id", orderID))
		return ErrNotFound
	}
	r.logger.InfoContext(ctx, "Order status updated",
		slog.Int64("order_id", orderID),
		slog.String("status", status),
	)
	return nil
}

// --- Update (обновление всего заказа) ---
func (r *orderRepo) Update(ctx context.Context, order *models.Order) error {
	r.logger.DebugContext(ctx, "Updating order",
		slog.Int64("order_id", order.ID),
	)

	query := `UPDATE orders SET 
		restaurant_id = $1, user_id = $2, address = $3, 
		total_price = $4, status = $5, estimated_delivery_at = $6, updated_at = NOW()
		WHERE id = $7`
	cmdTag, err := r.pool.Exec(ctx, query,
		order.RestaurantID, order.UserID, order.Address,
		order.TotalPrice, order.Status, order.EstimatedDeliveryAt,
		order.ID,
	)
	if err != nil {
		r.logger.ErrorContext(ctx, "Failed to update order", slog.String("error", err.Error()))
		return fmt.Errorf("update order: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		r.logger.WarnContext(ctx, "Order not found for update", slog.Int64("order_id", order.ID))
		return ErrNotFound
	}
	r.logger.InfoContext(ctx, "Order updated", slog.Int64("order_id", order.ID))
	return nil
}
