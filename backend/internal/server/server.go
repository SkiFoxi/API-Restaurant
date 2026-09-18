// Server это часть слоя хендлера
package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"avito-kitchen/backend/internal/handlers"
)

type Config struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func healthCheck(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Error("Failed to write response", "error", err)
		}
	}
}

// Run запускает HTTP-сервер с graceful shutdown
// Полный аналог ListenAndServe но теперь его можно кастомизировать
func Run(cfg Config, mux *http.ServeMux, log *slog.Logger) error {
	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		log.Info("Сервер запущен", slog.String("addr", cfg.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("SERVER Ошибка запуска сервера", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()
	// Ловим системный сигнал на завершение в канал quit
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("SERVER Получен сигнал завершения, останавливаем сервер...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Error("SERVER Ошибка при graceful shutdown", slog.String("error", err.Error()))
		return err
	}
	log.Info("SERVER Сервер остановлен")
	return nil
}

// SetupRouter создаёт и настраивает роутер
// Принимает хендлеры для регистрации всех эндпоинтов
func SetupRouter(
	orderHandler *handlers.OrderHandler,
	restaurantHandler *handlers.RestaurantHandler,
	menuHandler *handlers.MenuHandler,
	log *slog.Logger,
) *http.ServeMux {
	mux := http.NewServeMux()

	// Health check (публичный)
	mux.HandleFunc("GET /health", healthCheck(log))

	// Публичные эндпоинты (для пользователей)
	mux.HandleFunc("GET /api/v1/restaurants", restaurantHandler.ListRestaurants)
	mux.HandleFunc("GET /api/v1/restaurants/{id}", restaurantHandler.GetRestaurant)
	mux.HandleFunc("GET /api/v1/restaurants/{restaurantId}/menu", menuHandler.ListMenu)
	mux.HandleFunc("POST /api/v1/orders", orderHandler.CreateOrder)
	mux.HandleFunc("GET /api/v1/orders/{id}", orderHandler.GetOrder)
	mux.HandleFunc("DELETE /api/v1/orders/{id}", orderHandler.CancelOrder)

	// Внутренние эндпоинты (для ресторанов)
	mux.HandleFunc("GET /internal/orders", orderHandler.ListOrdersForRestaurant)
	mux.HandleFunc("PATCH /internal/orders/{id}/status", orderHandler.UpdateOrderStatus)
	mux.HandleFunc("POST /internal/menu", menuHandler.CreateMenuItem)
	mux.HandleFunc("PATCH /internal/menu/{id}/availability", menuHandler.UpdateAvailability)

	//Отдельно под микросервис "restaurant-demo"
	mux.HandleFunc("POST /internal/restaurants", restaurantHandler.CreateRestaurant)
	mux.HandleFunc("GET /internal/restaurants", restaurantHandler.FindRestaurantByName)

	return mux
}
