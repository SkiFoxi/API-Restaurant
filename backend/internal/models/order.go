// Структуры для репозиториев и сервисов
package models

import "time"

// Статусы заказа
const (
	OrderStatusNew        = "new"
	OrderStatusCollected  = "collected"
	OrderStatusCooking    = "cooking"
	OrderStatusReady      = "ready"
	OrderStatusDelivering = "delivering"
	OrderStatusDelivered  = "delivered"
	OrderStatusCancelled  = "cancelled"
)

// Определяет допустимые переходы статусов
// new -> cancelled (т.е. новый заказ можно отменить)
// cooking -> cancelled (а это отменить уже нельзя т.к. блюдо уже готовится)
var ValidStatusTransitions = map[string][]string{
	OrderStatusNew:        {OrderStatusCollected, OrderStatusCancelled},
	OrderStatusCollected:  {OrderStatusCooking, OrderStatusCancelled},
	OrderStatusCooking:    {OrderStatusReady},
	OrderStatusReady:      {OrderStatusDelivering},
	OrderStatusDelivering: {OrderStatusDelivered},
	OrderStatusDelivered:  {},
	OrderStatusCancelled:  {},
}

type Order struct {
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

// Залогировать ошибку этого метода нужно в основной программе!!!
// Проверяет, возможен ли переход в новый статус (использует ValidStatusTransitions)
func (o *Order) CanTransitionTo(newStatus string) bool {
	allowed, ok := ValidStatusTransitions[o.Status]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == newStatus {
			return true
		}
	}
	return false
}
