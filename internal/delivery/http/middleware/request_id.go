package middleware

import (
	"net/http"

	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/exPriceD/pr-reviewer-service/internal/infrastructure/logger"
)

// RequestID middleware извлекает request_id из Chi контекста и добавляет в logger контекст
// Chi middleware RequestID уже должен быть установлен до этого middleware
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := chimw.GetReqID(r.Context())

		if requestID != "" {
			w.Header().Set("X-Request-ID", requestID)
			ctx := logger.WithRequestID(r.Context(), requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			requestID = r.Header.Get("X-Request-ID")
			if requestID != "" {
				w.Header().Set("X-Request-ID", requestID)
				ctx := logger.WithRequestID(r.Context(), requestID)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				next.ServeHTTP(w, r)
			}
		}
	})
}
