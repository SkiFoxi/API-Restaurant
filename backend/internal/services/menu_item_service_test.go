package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"avito-kitchen/backend/internal/models"
	"avito-kitchen/backend/internal/repositories"
)

type mockMenuItemRepo struct {
	createFunc             func(ctx context.Context, item *models.MenuItem) error
	getByIDFunc            func(ctx context.Context, id int64) (*models.MenuItem, error)
	listByRestaurantFunc   func(ctx context.Context, restaurantID int64, availableOnly bool) ([]*models.MenuItem, error)
	updateAvailabilityFunc func(ctx context.Context, id int64, available bool) error
}

func (m *mockMenuItemRepo) Create(ctx context.Context, item *models.MenuItem) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, item)
	}
	return nil
}
func (m *mockMenuItemRepo) GetByID(ctx context.Context, id int64) (*models.MenuItem, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, repositories.ErrNotFound
}
func (m *mockMenuItemRepo) ListByRestaurant(ctx context.Context, restaurantID int64, availableOnly bool) ([]*models.MenuItem, error) {
	if m.listByRestaurantFunc != nil {
		return m.listByRestaurantFunc(ctx, restaurantID, availableOnly)
	}
	return nil, nil
}
func (m *mockMenuItemRepo) UpdateAvailability(ctx context.Context, id int64, available bool) error {
	if m.updateAvailabilityFunc != nil {
		return m.updateAvailabilityFunc(ctx, id, available)
	}
	return nil
}

func TestMenuItemService_CreateMenuItem_Success(t *testing.T) {
	var captured *models.MenuItem
	repo := &mockMenuItemRepo{
		createFunc: func(ctx context.Context, item *models.MenuItem) error {
			captured = item
			item.ID = 10
			return nil
		},
	}
	logger := testLogger()
	svc := NewMenuItemService(repo, logger)

	item, err := svc.CreateMenuItem(context.Background(), 1, "Pizza", "Tasty", 100)

	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, int64(10), item.ID)
	assert.Equal(t, "Pizza", item.Name)
	assert.True(t, item.IsAvailable)
	assert.NotNil(t, captured)
	assert.Equal(t, "Pizza", captured.Name)
}

func TestMenuItemService_CreateMenuItem_Error(t *testing.T) {
	repo := &mockMenuItemRepo{
		createFunc: func(ctx context.Context, item *models.MenuItem) error {
			return errors.New("db error")
		},
	}
	logger := testLogger()
	svc := NewMenuItemService(repo, logger)

	item, err := svc.CreateMenuItem(context.Background(), 1, "Pizza", "Tasty", 100)

	assert.Nil(t, item)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create menu item")
}

func TestMenuItemService_GetMenuItemByID_Success(t *testing.T) {
	expected := &models.MenuItem{ID: 1, Name: "Pizza"}
	repo := &mockMenuItemRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.MenuItem, error) {
			if id == 1 {
				return expected, nil
			}
			return nil, repositories.ErrNotFound
		},
	}
	logger := testLogger()
	svc := NewMenuItemService(repo, logger)

	item, err := svc.GetMenuItemByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, expected, item)
}

func TestMenuItemService_GetMenuItemByID_NotFound(t *testing.T) {
	repo := &mockMenuItemRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.MenuItem, error) {
			return nil, repositories.ErrNotFound
		},
	}
	logger := testLogger()
	svc := NewMenuItemService(repo, logger)

	item, err := svc.GetMenuItemByID(context.Background(), 999)

	assert.Nil(t, item)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, repositories.ErrNotFound))
}

func TestMenuItemService_ListMenuByRestaurant(t *testing.T) {
	expected := []*models.MenuItem{{ID: 1}, {ID: 2}}
	repo := &mockMenuItemRepo{
		listByRestaurantFunc: func(ctx context.Context, restaurantID int64, availableOnly bool) ([]*models.MenuItem, error) {
			return expected, nil
		},
	}
	logger := testLogger()
	svc := NewMenuItemService(repo, logger)

	items, err := svc.ListMenuByRestaurant(context.Background(), 1, false)

	assert.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestMenuItemService_ListMenuByRestaurant_Empty(t *testing.T) {
	repo := &mockMenuItemRepo{
		listByRestaurantFunc: func(ctx context.Context, restaurantID int64, availableOnly bool) ([]*models.MenuItem, error) {
			return []*models.MenuItem{}, nil
		},
	}
	logger := testLogger()
	svc := NewMenuItemService(repo, logger)

	items, err := svc.ListMenuByRestaurant(context.Background(), 1, true)

	assert.NoError(t, err)
	assert.Empty(t, items)
}

func TestMenuItemService_UpdateAvailability_Success(t *testing.T) {
	var updatedID int64
	var updatedAvailable bool
	repo := &mockMenuItemRepo{
		updateAvailabilityFunc: func(ctx context.Context, id int64, available bool) error {
			updatedID = id
			updatedAvailable = available
			return nil
		},
	}
	logger := testLogger()
	svc := NewMenuItemService(repo, logger)

	err := svc.UpdateAvailability(context.Background(), 5, false)

	assert.NoError(t, err)
	assert.Equal(t, int64(5), updatedID)
	assert.False(t, updatedAvailable)
}

func TestMenuItemService_UpdateAvailability_Error(t *testing.T) {
	repo := &mockMenuItemRepo{
		updateAvailabilityFunc: func(ctx context.Context, id int64, available bool) error {
			return errors.New("db error")
		},
	}
	logger := testLogger()
	svc := NewMenuItemService(repo, logger)

	err := svc.UpdateAvailability(context.Background(), 5, false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
