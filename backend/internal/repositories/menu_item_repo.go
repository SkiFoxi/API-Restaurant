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

type menuItemRepo struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewMenuItemRepository(pool *pgxpool.Pool, logger *slog.Logger) MenuItemRepository {
	return &menuItemRepo{pool: pool, logger: logger}
}

// --- GetByID ---
func (r *menuItemRepo) GetByID(ctx context.Context, id int64) (*models.MenuItem, error) {
	r.logger.DebugContext(ctx, "Fetching menu item by ID", slog.Int64("id", id))

	var item models.MenuItem
	query := `SELECT id, restaurant_id, name, description, price, is_available, created_at
		FROM menu_items WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&item.ID, &item.RestaurantID, &item.Name, &item.Description,
		&item.Price, &item.IsAvailable, &item.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.WarnContext(ctx, "Menu item not found", slog.Int64("id", id))
			return nil, ErrNotFound
		}
		r.logger.ErrorContext(ctx, "Failed to get menu item", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get menu item by id: %w", err)
	}
	return &item, nil
}

// --- ListByRestaurant ---
func (r *menuItemRepo) ListByRestaurant(ctx context.Context, restaurantID int64, availableOnly bool) ([]*models.MenuItem, error) {
	r.logger.DebugContext(ctx, "Listing menu items for restaurant",
		slog.Int64("restaurant_id", restaurantID),
		slog.Bool("available_only", availableOnly),
	)

	query := `SELECT id, restaurant_id, name, description, price, is_available, created_at
		FROM menu_items WHERE restaurant_id = $1`
	args := []any{restaurantID}
	if availableOnly {
		query += " AND is_available = true"
	}
	query += " ORDER BY id ASC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		r.logger.ErrorContext(ctx, "Failed to list menu items", slog.String("error", err.Error()))
		return nil, fmt.Errorf("list menu items: %w", err)
	}
	defer rows.Close()

	var items []*models.MenuItem
	for rows.Next() {
		var mi models.MenuItem
		err := rows.Scan(&mi.ID, &mi.RestaurantID, &mi.Name, &mi.Description,
			&mi.Price, &mi.IsAvailable, &mi.CreatedAt)
		if err != nil {
			r.logger.ErrorContext(ctx, "Failed to scan menu item row", slog.String("error", err.Error()))
			return nil, fmt.Errorf("scan menu item: %w", err)
		}
		items = append(items, &mi)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return items, nil
}

// --- Create ---
func (r *menuItemRepo) Create(ctx context.Context, item *models.MenuItem) error {
	r.logger.DebugContext(ctx, "Creating menu item",
		slog.Int64("restaurant_id", item.RestaurantID),
		slog.String("name", item.Name),
	)

	query := `INSERT INTO menu_items 
		(restaurant_id, name, description, price, is_available)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`
	err := r.pool.QueryRow(ctx, query,
		item.RestaurantID, item.Name, item.Description,
		item.Price, item.IsAvailable,
	).Scan(&item.ID, &item.CreatedAt)
	if err != nil {
		r.logger.ErrorContext(ctx, "Failed to create menu item", slog.String("error", err.Error()))
		return fmt.Errorf("create menu item: %w", err)
	}
	r.logger.InfoContext(ctx, "Menu item created", slog.Int64("id", item.ID))
	return nil
}

// --- UpdateAvailability ---
func (r *menuItemRepo) UpdateAvailability(ctx context.Context, id int64, available bool) error {
	r.logger.DebugContext(ctx, "Updating menu item availability",
		slog.Int64("id", id),
		slog.Bool("available", available),
	)

	query := `UPDATE menu_items SET is_available = $1 WHERE id = $2`
	cmdTag, err := r.pool.Exec(ctx, query, available, id)
	if err != nil {
		r.logger.ErrorContext(ctx, "Failed to update availability", slog.String("error", err.Error()))
		return fmt.Errorf("update availability: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		r.logger.WarnContext(ctx, "Menu item not found for availability update", slog.Int64("id", id))
		return ErrNotFound
	}
	r.logger.InfoContext(ctx, "Menu item availability updated",
		slog.Int64("id", id),
		slog.Bool("is_available", available),
	)
	return nil
}
