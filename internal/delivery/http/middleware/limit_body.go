package middleware

import (
	"net/http"
)

// LimitBodySize middleware ограничивает размер тела запроса
func LimitBodySize(maxBodySize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

			next.ServeHTTP(w, r)
		})
	}
}
