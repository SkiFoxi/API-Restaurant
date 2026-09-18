// В этом DTO используется теги из библиотеки для валидации.
// Я это сделал для наглядности(видны ограничения по полям)
package dto

import "time"

// Запрос на создание заказа (публичный)
type CreateOrderRequest struct {
	//Проверяем поля валидатором (required - поле с обязательным заполнением)
	RestaurantID int64  `json:"restaurant_id" validate:"required"`
	Address      string `json:"address" validate:"required,max=500"`
	Items        []struct {
		MenuItemID int64 `json:"menu_item_id" validate:"required"`
		Quantity   int   `json:"quantity" validate:"required,min=1"`
	} `json:"items" validate:"required,min=1"`
}

// Запрос на обновление статуса заказа (внутренний)
type UpdateOrderStatusRequest struct {
	// oneof сравнивает присланное значение с списком
	Status string `json:"status" validate:"required,oneof=new collected cooking ready delivering delivered cancelled"`
}

// Ответ для позиции заказа
type OrderItemResponse struct {
	MenuItemID  int64   `json:"menu_item_id"`
	Quantity    int     `json:"quantity"`
	PriceAtTime float64 `json:"price_at_time"`
}

// Ответ с данными заказа (без позиций – для MVP)
// Если нужны позиции, можно добавить поле Items []OrderItemResponse
type OrderResponse struct {
	ID                  int64      `json:"id"`
	RestaurantID        int64      `json:"restaurant_id"`
	UserID              int64      `json:"user_id"`
	Address             string     `json:"address"`
	TotalPrice          float64    `json:"total_price"`
	Status              string     `json:"status"`
	EstimatedDeliveryAt *time.Time `json:"estimated_delivery_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// Ответ для списка заказов (ресторан) – можно использовать тот же OrderResponse
