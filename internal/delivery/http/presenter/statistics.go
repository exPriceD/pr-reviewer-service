package presenter

import (
	"net/http"

	"github.com/exPriceD/pr-reviewer-service/internal/usecase/dto"
)

// RespondStatistics отправляет статистику в формате API
func RespondStatistics(w http.ResponseWriter, statusCode int, stats *dto.StatisticsDTO) {
	if stats == nil {
		RespondError(w, http.StatusInternalServerError, ErrorCodeInternalError, "statistics data is nil")
		return
	}
	RespondJSON(w, statusCode, stats)
}
