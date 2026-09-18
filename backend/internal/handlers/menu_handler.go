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

type MenuHandler struct {
	service *services.MenuItemService
	logger  *slog.Logger
}

func NewMenuHandler(service *services.MenuItemService, logger *slog.Logger) *MenuHandler {
	return &MenuHandler{service: service, logger: logger}
}

func (h *MenuHandler) ListMenu(w http.ResponseWriter, r *http.Request) {
	restaurantIDStr := r.PathValue("restaurantId")
	restaurantID, err := strconv.ParseInt(restaurantIDStr, 10, 64)
	if err != nil || restaurantID <= 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_RESTAURANT_ID", "restaurant id must be positive integer")
		return
	}
	availableOnly := r.URL.Query().Get("is_available") == "true"

	items, err := h.service.ListMenuByRestaurant(r.Context(), restaurantID, availableOnly)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "RESTAURANT_NOT_FOUND", "restaurant not found")
			return
		}
		h.logger.Error("Failed to list menu", "error", err)
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list menu")
		return
	}

	resp := mapper.ToMenuItemsResponse(items)
	WriteJSON(w, http.StatusOK, resp)
}

func (h *MenuHandler) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateMenuItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid JSON", "error", err)
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid JSON")
		return
	}
	if req.RestaurantID <= 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_RESTAURANT_ID", "restaurant_id must be positive")
		return
	}
	if req.Name == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_NAME", "name is required")
		return
	}
	if req.Price < 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_PRICE", "price must be >= 0")
		return
	}

	item, err := h.service.CreateMenuItem(r.Context(), req.RestaurantID, req.Name, req.Description, req.Price)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "RESTAURANT_NOT_FOUND", "restaurant not found")
			return
		}
		h.logger.Error("Failed to create menu item", "error", err)
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create menu item")
		return
	}

	resp := mapper.ToMenuItemResponse(item)
	WriteJSON(w, http.StatusCreated, resp)
}

func (h *MenuHandler) UpdateAvailability(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_ID", "menu item id must be positive")
		return
	}

	var req dto.UpdateAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid JSON")
		return
	}

	if err := h.service.UpdateAvailability(r.Context(), id, req.Available); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "menu item not found")
			return
		}
		h.logger.Error("Failed to update availability", "error", err)
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update availability")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"message": "availability updated"})
}
