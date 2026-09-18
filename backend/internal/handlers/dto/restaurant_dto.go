package dto

import "time"

// Запрос на создание ресторана (внутренний, для демо-сервиса)
type CreateRestaurantRequest struct {
	Name        string `json:"name" validate:"required"`
	Address     string `json:"address" validate:"required"`
	Phone       string `json:"phone"`
	OpeningTime string `json:"opening_time" validate:"required"` // "09:00:00"
	ClosingTime string `json:"closing_time" validate:"required"` // "22:00:00"
}

// Ответ с данными ресторана
type RestaurantResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	Phone       string    `json:"phone,omitempty"`
	OpeningTime string    `json:"opening_time"`
	ClosingTime string    `json:"closing_time"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}
