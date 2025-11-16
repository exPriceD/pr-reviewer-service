package middleware

import (
	"net/http"
	"time"

	"github.com/exPriceD/pr-reviewer-service/internal/domain/logger"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// Logger middleware для логирования HTTP запросов
func Logger(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := newResponseWriter(w)

			ctxLog := log.WithContext(r.Context())

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)

			logFields := []any{
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"bytes", wrapped.written,
			}

			if r.URL.RawQuery != "" {
				logFields = append(logFields, "query", r.URL.RawQuery)
			}

			switch {
			case wrapped.statusCode >= http.StatusInternalServerError:
				ctxLog.Error("HTTP request", logFields...)
			case wrapped.statusCode >= http.StatusBadRequest:
				ctxLog.Warn("HTTP request", logFields...)
			default:
				ctxLog.Info("HTTP request", logFields...)
			}
		})
	}
}
