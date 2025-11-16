package presenter

import (
	"net/http"

	"github.com/exPriceD/pr-reviewer-service/internal/usecase/dto"
)

// RespondUser отправляет пользователя в формате API
func RespondUser(w http.ResponseWriter, statusCode int, user *dto.UserDTO) {
	if user == nil {
		RespondError(w, http.StatusInternalServerError, ErrorCodeInternalError, "user data is nil")
		return
	}
	RespondJSON(w, statusCode, map[string]*dto.UserDTO{
		"user": user,
	})
}

// RespondUserReviews отправляет список PR пользователя в формате API
func RespondUserReviews(w http.ResponseWriter, statusCode int, userID string, prs []dto.PullRequestShortDTO) {
	if prs == nil {
		prs = []dto.PullRequestShortDTO{}
	}
	RespondJSON(w, statusCode, map[string]interface{}{
		"user_id":       userID,
		"pull_requests": prs,
	})
}
