package logger

import (
	"context"
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

// FileConfig – настройки записи в файл с ротацией
type FileConfig struct {
	Path       string // путь к файлу лога
	MaxSize    int    // максимальный размер файла в мегабайтах
	MaxBackups int    // максимальное количество старых файлов
	MaxAge     int    // максимальный возраст файлов в днях
	Compress   bool   // сжимать архивированные файлы
}

// Config – общие настройки логгера
type Config struct {
	Level      slog.Level
	JSON       bool
	AddContext bool
	File       *FileConfig // если nil – пишем в stdout
}

// LogContext – структура с полями для обогащения логов
type LogContext struct {
	UserID    int64
	RequestID string
	OrderID   int64
}

type contextKey struct{}

var ctxKey = contextKey{}

func WithContext(ctx context.Context, lc LogContext) context.Context {
	return context.WithValue(ctx, ctxKey, lc)
}

func ContextFrom(ctx context.Context) LogContext {
	if lc, ok := ctx.Value(ctxKey).(LogContext); ok {
		return lc
	}
	return LogContext{}
}

// contextHandler – middleware для добавления полей из контекста
type contextHandler struct {
	next slog.Handler
}

func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	lc := ContextFrom(ctx)
	if lc.UserID != 0 {
		r.AddAttrs(slog.Int64("user_id", lc.UserID))
	}
	if lc.RequestID != "" {
		r.AddAttrs(slog.String("request_id", lc.RequestID))
	}
	if lc.OrderID != 0 {
		r.AddAttrs(slog.Int64("order_id", lc.OrderID))
	}
	return h.next.Handle(ctx, r)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{next: h.next.WithAttrs(attrs)}
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{next: h.next.WithGroup(name)}
}

// New создаёт логгер с возможностью записи в файл с ротацией
func New(cfg Config) *slog.Logger {
	var writer io.Writer = os.Stdout

	if cfg.File != nil {
		writer = &lumberjack.Logger{
			Filename:   cfg.File.Path,
			MaxSize:    cfg.File.MaxSize, // мегабайты
			MaxBackups: cfg.File.MaxBackups,
			MaxAge:     cfg.File.MaxAge, // дни
			Compress:   cfg.File.Compress,
		}
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: cfg.Level}

	if cfg.JSON {
		handler = slog.NewJSONHandler(writer, opts)
	} else {
		handler = slog.NewTextHandler(writer, opts)
	}

	if cfg.AddContext {
		handler = &contextHandler{next: handler}
	}

	return slog.New(handler)
}

// Default – логгер по умолчанию (текстовый, контекстный, в stdout)
func Default() *slog.Logger {
	return New(Config{
		Level:      slog.LevelInfo,
		JSON:       false,
		AddContext: true,
	})
}
