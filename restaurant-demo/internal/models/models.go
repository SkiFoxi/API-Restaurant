package models

type RestaurantRegistration struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	Phone       string `json:"phone"`
	OpeningTime string `json:"opening_time"`
	ClosingTime string `json:"closing_time"`
}

type MenuItem struct {
	RestaurantID int64   `json:"restaurant_id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
}

type Order struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

type OrderListResponse struct {
	Items []Order `json:"items"`
}