package services

import (
	"context"
	"fmt"
	"log/slog"

	"avito-kitchen/backend/internal/models"
	"avito-kitchen/backend/internal/repositories"
)

type RestaurantService struct {
	//Тут тоже есть вшитый логер для действий БД
	repo repositories.RestaurantRepository
	//А это другой логер т.е. другой уровень логирования бизнес-событий
	logger *slog.Logger
}

func NewRestaurantService(repo repositories.RestaurantRepository, logger *slog.Logger) *RestaurantService {
	return &RestaurantService{
		repo:   repo,
		logger: logger,
	}
}

// RegisterRestaurant – регистрация нового ресторана (для внутреннего использования)
func (s *RestaurantService) RegisterRestaurant(ctx context.Context, name, address, phone, openingTime, closingTime string) (*models.Restaurant, error) {
	s.logger.InfoContext(ctx, "Registering restaurant", slog.String("name", name))

	rest := &models.Restaurant{
		Name:        name,
		Address:     address,
		Phone:       phone,
		OpeningTime: openingTime,
		ClosingTime: closingTime,
		IsActive:    true,
	}

	if err := s.repo.Create(ctx, rest); err != nil {
		return nil, fmt.Errorf("failed to register restaurant: %w", err)
	}
	return rest, nil
}

// GetRestaurantByID – получить ресторан по ID
func (s *RestaurantService) GetRestaurantByID(ctx context.Context, id int64) (*models.Restaurant, error) {
	s.logger.DebugContext(ctx, "Getting restaurant by ID", slog.Int64("id", id))
	return s.repo.GetByID(ctx, id)
}

// FindRestaurantByName – найти ресторан по точному имени (используется demo-сервисом,
// чтобы не регистрировать ресторан заново при каждом рестарте)
func (s *RestaurantService) FindRestaurantByName(ctx context.Context, name string) (*models.Restaurant, error) {
	s.logger.DebugContext(ctx, "Finding restaurant by name", slog.String("name", name))
	return s.repo.GetByName(ctx, name)
}

// ListRestaurants – список ресторанов с пагинацией
func (s *RestaurantService) ListRestaurants(ctx context.Context, limit int, cursor int64) ([]*models.Restaurant, int64, error) {
	s.logger.DebugContext(ctx, "Listing restaurants", slog.Int("limit", limit), slog.Int64("cursor", cursor))
	return s.repo.List(ctx, limit, cursor)
}
