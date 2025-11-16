package presenter

import (
	"net/http"

	"github.com/exPriceD/pr-reviewer-service/internal/usecase/dto"
)

// RespondPullRequest отправляет PR в формате API
func RespondPullRequest(w http.ResponseWriter, statusCode int, pr *dto.PullRequestDTO) {
	if pr == nil {
		RespondError(w, http.StatusInternalServerError, ErrorCodeInternalError, "pull request data is nil")
		return
	}
	RespondJSON(w, statusCode, map[string]*dto.PullRequestDTO{
		"pr": pr,
	})
}

// RespondPullRequestReassign отправляет результат переназначения ревьювера
func RespondPullRequestReassign(w http.ResponseWriter, statusCode int, pr *dto.PullRequestDTO, replacedBy string) {
	if pr == nil {
		RespondError(w, http.StatusInternalServerError, ErrorCodeInternalError, "pull request data is nil")
		return
	}
	if replacedBy == "" {
		RespondError(w, http.StatusInternalServerError, ErrorCodeInternalError, "replaced_by is empty")
		return
	}
	RespondJSON(w, statusCode, map[string]interface{}{
		"pr":          pr,
		"replaced_by": replacedBy,
	})
}
