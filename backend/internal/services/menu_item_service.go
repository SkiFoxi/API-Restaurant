package services

import (
	"context"
	"fmt"
	"log/slog"

	"avito-kitchen/backend/internal/models"
	"avito-kitchen/backend/internal/repositories"
)

type MenuItemService struct {
	repo   repositories.MenuItemRepository
	logger *slog.Logger
}

func NewMenuItemService(repo repositories.MenuItemRepository, logger *slog.Logger) *MenuItemService {
	return &MenuItemService{
		repo:   repo,
		logger: logger,
	}
}

// CreateMenuItem – добавить позицию в меню (для внутреннего использования)
func (s *MenuItemService) CreateMenuItem(ctx context.Context, restaurantID int64, name, description string, price float64) (*models.MenuItem, error) {
	s.logger.InfoContext(ctx, "Creating menu item",
		slog.Int64("restaurant_id", restaurantID),
		slog.String("name", name),
	)

	item := &models.MenuItem{
		RestaurantID: restaurantID,
		Name:         name,
		Description:  description,
		Price:        price,
		IsAvailable:  true,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to create menu item: %w", err)
	}
	return item, nil
}

// GetMenuItemByID – получить позицию по ID
func (s *MenuItemService) GetMenuItemByID(ctx context.Context, id int64) (*models.MenuItem, error) {
	return s.repo.GetByID(ctx, id)
}

// ListMenuByRestaurant – меню ресторана (опционально только доступные)
func (s *MenuItemService) ListMenuByRestaurant(ctx context.Context, restaurantID int64, availableOnly bool) ([]*models.MenuItem, error) {
	s.logger.DebugContext(ctx, "Listing menu items for restaurant",
		slog.Int64("restaurant_id", restaurantID),
		slog.Bool("available_only", availableOnly),
	)
	return s.repo.ListByRestaurant(ctx, restaurantID, availableOnly)
}

// UpdateAvailability – включить/выключить позицию
func (s *MenuItemService) UpdateAvailability(ctx context.Context, id int64, available bool) error {
	s.logger.InfoContext(ctx, "Updating menu item availability",
		slog.Int64("id", id),
		slog.Bool("available", available),
	)
	return s.repo.UpdateAvailability(ctx, id, available)
}
