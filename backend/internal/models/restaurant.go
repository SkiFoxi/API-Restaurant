// Структура для репозиториев и сервисов
package models

import "time"

type Restaurant struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	Phone       string    `json:"phone,omitempty"`
	OpeningTime string    `json:"opening_time"` // "09:00:00"
	ClosingTime string    `json:"closing_time"` // "22:00:00"
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}
