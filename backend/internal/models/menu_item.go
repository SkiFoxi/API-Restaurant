// Структура для репозиториев и сервисов
package models

import "time"

type MenuItem struct {
	ID           int64     `json:"id"`
	RestaurantID int64     `json:"restaurant_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	Price        float64   `json:"price"`
	IsAvailable  bool      `json:"is_available"`
	CreatedAt    time.Time `json:"created_at"`
}
