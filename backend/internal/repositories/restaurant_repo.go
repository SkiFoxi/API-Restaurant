// Реализация интерфейса RestaurantRepository
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

type restaurantRepo struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewRestaurantRepository(pool *pgxpool.Pool, logger *slog.Logger) RestaurantRepository {
	return &restaurantRepo{pool: pool, logger: logger}
}

// --- GetByID ---
func (r *restaurantRepo) GetByID(ctx context.Context, id int64) (*models.Restaurant, error) {
	r.logger.DebugContext(ctx, "Fetching restaurant by ID", slog.Int64("id", id))

	var rest models.Restaurant
	query := `SELECT id, name, address, phone, opening_time, closing_time, is_active, created_at
		FROM restaurants WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&rest.ID, &rest.Name, &rest.Address, &rest.Phone,
		&rest.OpeningTime, &rest.ClosingTime, &rest.IsActive, &rest.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.WarnContext(ctx, "Restaurant not found", slog.Int64("id", id))
			return nil, ErrNotFound
		}
		r.logger.ErrorContext(ctx, "Failed to get restaurant", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get restaurant by id: %w", err)
	}
	return &rest, nil
}

// --- GetByName ---
func (r *restaurantRepo) GetByName(ctx context.Context, name string) (*models.Restaurant, error) {
	r.logger.DebugContext(ctx, "Fetching restaurant by name", slog.String("name", name))

	var rest models.Restaurant
	query := `SELECT id, name, address, phone, opening_time, closing_time, is_active, created_at
		FROM restaurants WHERE name = $1 ORDER BY id ASC LIMIT 1`
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&rest.ID, &rest.Name, &rest.Address, &rest.Phone,
		&rest.OpeningTime, &rest.ClosingTime, &rest.IsActive, &rest.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.DebugContext(ctx, "Restaurant not found by name", slog.String("name", name))
			return nil, ErrNotFound
		}
		r.logger.ErrorContext(ctx, "Failed to get restaurant by name", slog.String("error", err.Error()))
		return nil, fmt.Errorf("get restaurant by name: %w", err)
	}
	return &rest, nil
}

// --- Create ---
func (r *restaurantRepo) Create(ctx context.Context, restaurant *models.Restaurant) error {
	r.logger.DebugContext(ctx, "Creating restaurant", slog.String("name", restaurant.Name))

	query := `INSERT INTO restaurants 
		(name, address, phone, opening_time, closing_time, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`
	err := r.pool.QueryRow(ctx, query,
		restaurant.Name, restaurant.Address, restaurant.Phone,
		restaurant.OpeningTime, restaurant.ClosingTime, restaurant.IsActive,
	).Scan(&restaurant.ID, &restaurant.CreatedAt)
	if err != nil {
		r.logger.ErrorContext(ctx, "Failed to create restaurant", slog.String("error", err.Error()))
		return fmt.Errorf("create restaurant: %w", err)
	}
	r.logger.InfoContext(ctx, "Restaurant created", slog.Int64("id", restaurant.ID))
	return nil
}

// --- List (только пагинация, без фильтра) ---
func (r *restaurantRepo) List(ctx context.Context, limit int, cursor int64) ([]*models.Restaurant, int64, error) {
	r.logger.DebugContext(ctx, "Listing restaurants",
		slog.Int("limit", limit),
		slog.Int64("cursor", cursor),
	)

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `SELECT id, name, address, phone, opening_time, closing_time, is_active, created_at
		FROM restaurants
		WHERE id > $1
		ORDER BY id ASC
		LIMIT $2`

	rows, err := r.pool.Query(ctx, query, cursor, limit)
	if err != nil {
		r.logger.ErrorContext(ctx, "Failed to list restaurants", slog.String("error", err.Error()))
		return nil, 0, fmt.Errorf("list restaurants: %w", err)
	}
	defer rows.Close()

	var restaurants []*models.Restaurant
	var lastID int64
	for rows.Next() {
		var rest models.Restaurant
		err := rows.Scan(&rest.ID, &rest.Name, &rest.Address, &rest.Phone,
			&rest.OpeningTime, &rest.ClosingTime, &rest.IsActive, &rest.CreatedAt)
		if err != nil {
			r.logger.ErrorContext(ctx, "Failed to scan restaurant row", slog.String("error", err.Error()))
			return nil, 0, fmt.Errorf("scan restaurant: %w", err)
		}
		restaurants = append(restaurants, &rest)
		lastID = rest.ID
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration: %w", err)
	}

	var nextCursor int64
	if len(restaurants) == limit {
		nextCursor = lastID
	} else {
		nextCursor = 0
	}

	return restaurants, nextCursor, nil
}
