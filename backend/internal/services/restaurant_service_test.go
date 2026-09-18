package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"avito-kitchen/backend/internal/models"
	"avito-kitchen/backend/internal/repositories"
)

type mockRestaurantRepo struct {
	createFunc    func(ctx context.Context, restaurant *models.Restaurant) error
	getByIDFunc   func(ctx context.Context, id int64) (*models.Restaurant, error)
	getByNameFunc func(ctx context.Context, name string) (*models.Restaurant, error)
	listFunc      func(ctx context.Context, limit int, cursor int64) ([]*models.Restaurant, int64, error)
}

func (m *mockRestaurantRepo) Create(ctx context.Context, restaurant *models.Restaurant) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, restaurant)
	}
	return nil
}
func (m *mockRestaurantRepo) GetByID(ctx context.Context, id int64) (*models.Restaurant, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, repositories.ErrNotFound
}
func (m *mockRestaurantRepo) GetByName(ctx context.Context, name string) (*models.Restaurant, error) {
	if m.getByNameFunc != nil {
		return m.getByNameFunc(ctx, name)
	}
	return nil, repositories.ErrNotFound
}
func (m *mockRestaurantRepo) List(ctx context.Context, limit int, cursor int64) ([]*models.Restaurant, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, limit, cursor)
	}
	return nil, 0, nil
}

func TestRestaurantService_RegisterRestaurant_Success(t *testing.T) {
	var captured *models.Restaurant
	repo := &mockRestaurantRepo{
		createFunc: func(ctx context.Context, rest *models.Restaurant) error {
			captured = rest
			rest.ID = 1
			return nil
		},
	}
	logger := testLogger()
	svc := NewRestaurantService(repo, logger)

	rest, err := svc.RegisterRestaurant(context.Background(), "Test", "Addr", "123", "09:00", "22:00")

	assert.NoError(t, err)
	assert.NotNil(t, rest)
	assert.Equal(t, int64(1), rest.ID)
	assert.Equal(t, "Test", rest.Name)
	assert.True(t, rest.IsActive)
	assert.NotNil(t, captured)
	assert.Equal(t, "Test", captured.Name)
}

func TestRestaurantService_GetRestaurantByID_Success(t *testing.T) {
	expected := &models.Restaurant{ID: 1, Name: "Test"}
	repo := &mockRestaurantRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Restaurant, error) {
			if id == 1 {
				return expected, nil
			}
			return nil, repositories.ErrNotFound
		},
	}
	logger := testLogger()
	svc := NewRestaurantService(repo, logger)

	rest, err := svc.GetRestaurantByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, expected, rest)
}

func TestRestaurantService_GetRestaurantByID_NotFound(t *testing.T) {
	repo := &mockRestaurantRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Restaurant, error) {
			return nil, repositories.ErrNotFound
		},
	}
	logger := testLogger()
	svc := NewRestaurantService(repo, logger)

	rest, err := svc.GetRestaurantByID(context.Background(), 999)

	assert.Nil(t, rest)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, repositories.ErrNotFound))
}

func TestRestaurantService_ListRestaurants(t *testing.T) {
	expected := []*models.Restaurant{{ID: 1}, {ID: 2}}
	repo := &mockRestaurantRepo{
		listFunc: func(ctx context.Context, limit int, cursor int64) ([]*models.Restaurant, int64, error) {
			return expected, 2, nil
		},
	}
	logger := testLogger()
	svc := NewRestaurantService(repo, logger)

	rests, next, err := svc.ListRestaurants(context.Background(), 10, 0)

	assert.NoError(t, err)
	assert.Len(t, rests, 2)
	assert.Equal(t, int64(2), next)
}

func TestRestaurantService_RegisterRestaurant_Error(t *testing.T) {
	repo := &mockRestaurantRepo{
		createFunc: func(ctx context.Context, rest *models.Restaurant) error {
			return errors.New("database error")
		},
	}
	logger := testLogger()
	svc := NewRestaurantService(repo, logger)

	rest, err := svc.RegisterRestaurant(context.Background(), "Test", "Addr", "123", "09:00", "22:00")

	assert.Nil(t, rest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to register restaurant")
}

func TestRestaurantService_FindRestaurantByName_Success(t *testing.T) {
	expected := &models.Restaurant{ID: 1, Name: "Demo Restaurant"}
	repo := &mockRestaurantRepo{
		getByNameFunc: func(ctx context.Context, name string) (*models.Restaurant, error) {
			if name == "Demo Restaurant" {
				return expected, nil
			}
			return nil, repositories.ErrNotFound
		},
	}
	logger := testLogger()
	svc := NewRestaurantService(repo, logger)

	rest, err := svc.FindRestaurantByName(context.Background(), "Demo Restaurant")

	assert.NoError(t, err)
	assert.Equal(t, expected, rest)
}

func TestRestaurantService_FindRestaurantByName_NotFound(t *testing.T) {
	repo := &mockRestaurantRepo{
		getByNameFunc: func(ctx context.Context, name string) (*models.Restaurant, error) {
			return nil, repositories.ErrNotFound
		},
	}
	logger := testLogger()
	svc := NewRestaurantService(repo, logger)

	rest, err := svc.FindRestaurantByName(context.Background(), "Unknown")

	assert.Nil(t, rest)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, repositories.ErrNotFound))
}