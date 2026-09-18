package dto

import "time"

// Запрос на создание позиции меню (внутренний)
type CreateMenuItemRequest struct {
	RestaurantID int64   `json:"restaurant_id" validate:"required"`
	Name         string  `json:"name" validate:"required"`
	Description  string  `json:"description"`
	Price        float64 `json:"price" validate:"required,min=0"`
}

// Запрос на обновление доступности
type UpdateAvailabilityRequest struct {
	Available bool `json:"is_available" validate:"required"`
}

// Ответ с данными позиции меню
type MenuItemResponse struct {
	ID           int64     `json:"id"`
	RestaurantID int64     `json:"restaurant_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	Price        float64   `json:"price"`
	IsAvailable  bool      `json:"is_available"`
	CreatedAt    time.Time `json:"created_at"`
}
