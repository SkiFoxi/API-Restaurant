package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"avito-kitchen/backend/internal/database"
	"avito-kitchen/backend/internal/handlers"
	"avito-kitchen/backend/internal/logger"
	"avito-kitchen/backend/internal/repositories"
	"avito-kitchen/backend/internal/server"
	"avito-kitchen/backend/internal/services"
)

func main() {
	log := logger.New(logger.Config{
		Level:      slog.LevelInfo,
		JSON:       false,
		AddContext: true,
		File: &logger.FileConfig{
			Path:       "app.log",
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
		},
	})
	slog.SetDefault(log)

	//Добавил начальные значения для БД, чтобы легче было проверять проект
	//Они впишутся, если при подключении ничего не указать
	dbCfg := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "avito"),
		Password: getEnv("DB_PASSWORD", "avito_pass"),
		DBName:   getEnv("DB_NAME", "avito_kitchen"),
	}
	ctx := context.Background()
	pool, err := database.NewPool(ctx, dbCfg, log)
	if err != nil {
		log.Error("Критическая ошибка при подключении к БД", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	restRepo := repositories.NewRestaurantRepository(pool, log)
	menuRepo := repositories.NewMenuItemRepository(pool, log)
	orderRepo := repositories.NewOrderRepository(pool, log)

	restService := services.NewRestaurantService(restRepo, log)
	menuService := services.NewMenuItemService(menuRepo, log)
	orderService := services.NewOrderService(orderRepo, menuRepo, restRepo, log)

	restHandler := handlers.NewRestaurantHandler(restService, log)
	menuHandler := handlers.NewMenuHandler(menuService, log)
	orderHandler := handlers.NewOrderHandler(orderService, log)

	mux := server.SetupRouter(orderHandler, restHandler, menuHandler, log)

	//Минимальная защита от дудос атаки
	cfg := server.Config{
		Addr:         ":8080",          // порт, на котором слушает сервер
		ReadTimeout:  5 * time.Second,  // максимальное время чтения запроса клиента
		WriteTimeout: 10 * time.Second, // максимальное время на отправку ответа клиенту
		IdleTimeout:  15 * time.Second, // твремя жизни открытого соединения
	}
	if err := server.Run(cfg, mux, log); err != nil {
		log.Error("Ошибка работы сервера", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

/*
	 Если нет глобальных переменных под по названию в переменной key,
		то берется начальное значение которое вписано в переменную defaultValue
*/
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
