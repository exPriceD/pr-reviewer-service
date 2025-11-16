package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/exPriceD/pr-reviewer-service/internal/domain/logger"
)

var _ logger.Logger = (*SlogLogger)(nil)

// SlogLogger реализация Logger на основе slog
type SlogLogger struct {
	logger *slog.Logger
}

type Config struct {
	Level  string // "debug", "info", "warn", "error"
	Format string // "json", "text"
}

// NewSlogLogger создает новый slog логгер
func NewSlogLogger(config Config) logger.Logger {
	level := parseLevel(config.Level)

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug,
	}

	switch config.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return &SlogLogger{
		logger: slog.New(handler),
	}
}

func (l *SlogLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *SlogLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *SlogLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *SlogLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// With возвращает новый логгер с дополнительными полями
func (l *SlogLogger) With(args ...any) logger.Logger {
	return &SlogLogger{
		logger: l.logger.With(args...),
	}
}

// WithContext возвращает логгер с полями из контекста
func (l *SlogLogger) WithContext(ctx context.Context) logger.Logger {
	fields := ExtractFields(ctx)
	if len(fields) == 0 {
		return l
	}
	return &SlogLogger{
		logger: l.logger.With(fields...),
	}
}

// parseLevel парсинг и преобразование в slog.Level
func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
