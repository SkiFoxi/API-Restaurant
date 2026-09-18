package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config содержит настройки подключения к БД
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// NewPool создаёт пул соединений с PostgreSQL
func NewPool(ctx context.Context, cfg Config, logger *slog.Logger) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	)

	logger.Info("Подключение к PostgreSQL",
		slog.String("host", cfg.Host),
		slog.String("port", cfg.Port),
		slog.String("database", cfg.DBName),
	)

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		logger.Error("Не удалось создать пул соединений", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Проверка связи
	if err := pool.Ping(ctx); err != nil {
		logger.Error("Ping к БД не удался", slog.String("error", err.Error()))
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	logger.Info("Успешное подключение к PostgreSQL")
	return pool, nil
}
