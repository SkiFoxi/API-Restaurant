// Структура для репозиториев и сервисов
package models

type OrderItem struct {
	ID          int64   `json:"id"`
	OrderID     int64   `json:"order_id"`
	MenuItemID  int64   `json:"menu_item_id"`
	Quantity    int     `json:"quantity"`      //Количество товара в заказе
	PriceAtTime float64 `json:"price_at_time"` // Цена на момент заказа
}
