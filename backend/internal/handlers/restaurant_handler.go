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

type RestaurantHandler struct {
	service *services.RestaurantService
	logger  *slog.Logger
}

func NewRestaurantHandler(service *services.RestaurantService, logger *slog.Logger) *RestaurantHandler {
	return &RestaurantHandler{service: service, logger: logger}
}

func (h *RestaurantHandler) ListRestaurants(w http.ResponseWriter, r *http.Request) {
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

	restaurants, nextCursor, err := h.service.ListRestaurants(r.Context(), limit, cursor)
	if err != nil {
		h.logger.Error("Failed to list restaurants", "error", err)
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list restaurants")
		return
	}

	resp := mapper.ToRestaurantsResponse(restaurants)
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items":       resp,
		"next_cursor": nextCursor,
	})
}

func (h *RestaurantHandler) GetRestaurant(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_ID", "restaurant id must be positive integer")
		return
	}

	rest, err := h.service.GetRestaurantByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "restaurant not found")
			return
		}
		h.logger.Error("Failed to get restaurant", "error", err)
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get restaurant")
		return
	}

	resp := mapper.ToRestaurantResponse(rest)
	WriteJSON(w, http.StatusOK, resp)
}

// FindRestaurantByName – GET /internal/restaurants?name=... (используется demo-сервисом,
// чтобы не регистрировать ресторан заново при каждом рестарте)
func (h *RestaurantHandler) FindRestaurantByName(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_NAME", "name is required")
		return
	}

	rest, err := h.service.FindRestaurantByName(r.Context(), name)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "restaurant not found")
			return
		}
		h.logger.Error("Failed to find restaurant by name", "error", err)
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to find restaurant")
		return
	}

	resp := mapper.ToRestaurantResponse(rest)
	WriteJSON(w, http.StatusOK, resp)
}

// CreateRestaurant – POST /internal/restaurants (для регистрации нового заведения)
func (h *RestaurantHandler) CreateRestaurant(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateRestaurantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid JSON", "error", err)
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid JSON")
		return
	}
	if req.Name == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_NAME", "name is required")
		return
	}
	if req.Address == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_ADDRESS", "address is required")
		return
	}
	if req.OpeningTime == "" || req.ClosingTime == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_HOURS", "opening_time and closing_time are required")
		return
	}

	// Создаём ресторан через сервис
	rest, err := h.service.RegisterRestaurant(r.Context(), req.Name, req.Address, req.Phone, req.OpeningTime, req.ClosingTime)
	if err != nil {
		h.logger.Error("Failed to register restaurant", "error", err)
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to register restaurant")
		return
	}

	resp := mapper.ToRestaurantResponse(rest)
	WriteJSON(w, http.StatusCreated, resp)
}
