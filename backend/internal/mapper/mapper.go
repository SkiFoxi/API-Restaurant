// Не дает хендлеру прямо взаимодействовать с models и создает функции для переводо одного типа в другой
package mapper

import (
	"avito-kitchen/backend/internal/handlers/dto"
	"avito-kitchen/backend/internal/models"
)

func ToOrderResponse(order *models.Order) dto.OrderResponse {
	return dto.OrderResponse{
		ID:                  order.ID,
		RestaurantID:        order.RestaurantID,
		UserID:              order.UserID,
		Address:             order.Address,
		TotalPrice:          order.TotalPrice,
		Status:              order.Status,
		EstimatedDeliveryAt: order.EstimatedDeliveryAt,
		CreatedAt:           order.CreatedAt,
		UpdatedAt:           order.UpdatedAt,
	}
}

func ToOrdersResponse(orders []*models.Order) []dto.OrderResponse {
	resp := make([]dto.OrderResponse, len(orders))
	for i, o := range orders {
		resp[i] = ToOrderResponse(o)
	}
	return resp
}

func ToRestaurantResponse(rest *models.Restaurant) dto.RestaurantResponse {
	return dto.RestaurantResponse{
		ID:          rest.ID,
		Name:        rest.Name,
		Address:     rest.Address,
		Phone:       rest.Phone,
		OpeningTime: rest.OpeningTime,
		ClosingTime: rest.ClosingTime,
		IsActive:    rest.IsActive,
		CreatedAt:   rest.CreatedAt,
	}
}

func ToRestaurantsResponse(restaurants []*models.Restaurant) []dto.RestaurantResponse {
	resp := make([]dto.RestaurantResponse, len(restaurants))
	for i, r := range restaurants {
		resp[i] = ToRestaurantResponse(r)
	}
	return resp
}

func ToMenuItemResponse(item *models.MenuItem) dto.MenuItemResponse {
	return dto.MenuItemResponse{
		ID:           item.ID,
		RestaurantID: item.RestaurantID,
		Name:         item.Name,
		Description:  item.Description,
		Price:        item.Price,
		IsAvailable:  item.IsAvailable,
		CreatedAt:    item.CreatedAt,
	}
}

func ToMenuItemsResponse(items []*models.MenuItem) []dto.MenuItemResponse {
	resp := make([]dto.MenuItemResponse, len(items))
	for i, it := range items {
		resp[i] = ToMenuItemResponse(it)
	}
	return resp
}
