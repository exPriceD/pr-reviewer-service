package middleware

import (
	"net/http"

	"github.com/exPriceD/pr-reviewer-service/internal/domain/logger"
)

// Recovery middleware восстанавливает после panic
func Recovery(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrapped := &recoveryWriter{ResponseWriter: w}

			defer func() {
				if err := recover(); err != nil {
					ctxLog := log.WithContext(r.Context())
					ctxLog.Error("panic recovered",
						"error", err,
						"method", r.Method,
						"path", r.URL.Path,
					)

					if !wrapped.headerWritten {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusInternalServerError)
						wrapped.headerWritten = true
					}

					if _, writeErr := wrapped.Write([]byte(`{"error":{"code":"INTERNAL_ERROR","message":"internal server error"}}`)); writeErr != nil {
						ctxLog.Error("failed to write recovery response", "error", writeErr)
					}
				}
			}()

			next.ServeHTTP(wrapped, r)
		})
	}
}

// recoveryWriter отслеживает было ли WriteHeader вызван
type recoveryWriter struct {
	http.ResponseWriter
	headerWritten bool
}

func (rw *recoveryWriter) WriteHeader(statusCode int) {
	if !rw.headerWritten {
		rw.ResponseWriter.WriteHeader(statusCode)
		rw.headerWritten = true
	}
}
