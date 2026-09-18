package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"avito-kitchen/backend/internal/handlers/dto"
	"avito-kitchen/backend/internal/mapper"
	"avito-kitchen/backend/internal/repositories"
	"avito-kitchen/backend/internal/services"
)

type OrderHandler struct {
	service *services.OrderService
	logger  *slog.Logger
}

func NewOrderHandler(service *services.OrderService, logger *slog.Logger) *OrderHandler {
	return &OrderHandler{service: service, logger: logger}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid JSON", "error", err)
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
		return
	}

	if req.RestaurantID <= 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_RESTAURANT_ID", "restaurant_id must be positive")
		return
	}
	if req.Address == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_ADDRESS", "address is required")
		return
	}
	if len(req.Items) == 0 {
		WriteError(w, http.StatusBadRequest, "EMPTY_CART", "cart must contain at least one item")
		return
	}
	for _, item := range req.Items {
		if item.MenuItemID <= 0 {
			WriteError(w, http.StatusBadRequest, "INVALID_ITEM", "menu_item_id must be positive")
			return
		}
		if item.Quantity <= 0 {
			WriteError(w, http.StatusBadRequest, "INVALID_QUANTITY", "quantity must be positive")
			return
		}
	}

	cartItems := make([]struct {
		MenuItemID int64
		Quantity   int
	}, len(req.Items))
	for i, item := range req.Items {
		cartItems[i] = struct {
			MenuItemID int64
			Quantity   int
		}{
			MenuItemID: item.MenuItemID,
			Quantity:   item.Quantity,
		}
	}

	userID := int64(1) // заглушка

	order, err := h.service.CreateOrder(r.Context(), req.RestaurantID, userID, req.Address, cartItems)
	if err != nil {
		h.logger.Error("Failed to create order", "error", err)
		switch {
		case errors.Is(err, repositories.ErrNotFound):
			WriteError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		default:
			WriteError(w, http.StatusBadRequest, "ORDER_FAILED", err.Error())
		}
		return
	}

	resp := mapper.ToOrderResponse(order)
	WriteJSON(w, http.StatusCreated, resp)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_ID", "order id must be positive integer")
		return
	}

	order, err := h.service.GetOrderByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "order not found")
			return
		}
		h.logger.Error("Failed to get order", "error", err)
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get order")
		return
	}

	resp := mapper.ToOrderResponse(order)
	WriteJSON(w, http.StatusOK, resp)
}

func (h *OrderHandler) ListOrdersForRestaurant(w http.ResponseWriter, r *http.Request) {
	restaurantIDStr := r.URL.Query().Get("restaurant_id")
	if restaurantIDStr == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_RESTAURANT_ID", "restaurant_id is required")
		return
	}
	restaurantID, err := strconv.ParseInt(restaurantIDStr, 10, 64)
	if err != nil || restaurantID <= 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_RESTAURANT_ID", "restaurant_id must be positive")
		return
	}

	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	cursorStr := r.URL.Query().Get("cursor")
	cursor := int64(0)
	if cursorStr != "" {
		if v, err := strconv.ParseInt(cursorStr, 10, 64); err == nil && v > 0 {
			cursor = v
		}
	}

	orders, nextCursor, err := h.service.ListOrdersForRestaurant(r.Context(), restaurantID, status, cursor, limit)
	if err != nil {
		h.logger.Error("Failed to list orders", "error", err)
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list orders")
		return
	}

	resp := mapper.ToOrdersResponse(orders)
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items":       resp,
		"next_cursor": nextCursor,
	})
}

func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_ID", "order id must be positive")
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid JSON")
		return
	}
	if req.Status == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_STATUS", "status is required")
		return
	}

	if err := h.service.UpdateOrderStatus(r.Context(), id, req.Status); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "order not found")
			return
		}
		h.logger.Error("Failed to update status", "error", err)
		WriteError(w, http.StatusBadRequest, "INVALID_STATUS", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"message": "status updated"})
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_ID", "order id must be positive")
		return
	}

	if err := h.service.CancelOrder(r.Context(), id); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "order not found")
			return
		}
		h.logger.Error("Failed to cancel order", "error", err)
		WriteError(w, http.StatusBadRequest, "CANCEL_FAILED", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"message": "order cancelled"})
}
