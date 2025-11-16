package presenter

import (
	"net/http"

	"github.com/exPriceD/pr-reviewer-service/internal/usecase/dto"
)

// RespondTeam отправляет команду в формате API
func RespondTeam(w http.ResponseWriter, statusCode int, team *dto.TeamDTO) {
	if team == nil {
		RespondError(w, http.StatusInternalServerError, ErrorCodeInternalError, "team data is nil")
		return
	}
	RespondJSON(w, statusCode, map[string]*dto.TeamDTO{
		"team": team,
	})
}
