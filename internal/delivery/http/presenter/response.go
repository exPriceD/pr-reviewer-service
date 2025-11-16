package presenter

import (
	"encoding/json"
	"net/http"
)

// RespondJSON отправляет JSON ответ с указанным статус кодом
// ВАЖНО: WriteHeader должен быть вызван только один раз
func RespondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		_ = err
	}
}

// RespondError отправляет ошибку в формате API
func RespondError(w http.ResponseWriter, statusCode int, code, message string) {
	RespondJSON(w, statusCode, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

// RespondSuccess отправляет успешный ответ с данными
func RespondSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	RespondJSON(w, statusCode, data)
}
