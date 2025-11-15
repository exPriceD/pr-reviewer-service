package logger

import "context"

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
)

// WithRequestID добавляет request ID в контекст
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID извлекает request ID из контекста
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// GetUserID извлекает user ID из контекста
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}

// ExtractFields извлекает поля для логирования из контекста
func ExtractFields(ctx context.Context) []any {
	var fields []any

	if requestID := GetRequestID(ctx); requestID != "" {
		fields = append(fields, "request_id", requestID)
	}

	if userID := GetUserID(ctx); userID != "" {
		fields = append(fields, "user_id", userID)
	}

	return fields
}
