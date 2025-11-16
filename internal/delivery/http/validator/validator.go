package validator

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/exPriceD/pr-reviewer-service/internal/delivery/http/presenter"
	"github.com/exPriceD/pr-reviewer-service/internal/usecase/dto"
)

// ValidationError представляет ошибку валидации
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

// ValidateCreateTeamRequest валидирует CreateTeamRequest
func ValidateCreateTeamRequest(req dto.CreateTeamRequest) []ValidationError {
	var errors []ValidationError

	if req.TeamName == "" {
		errors = append(errors, ValidationError{
			Field:   "team_name",
			Message: "team_name is required",
		})
	}

	if len(req.Members) == 0 {
		errors = append(errors, ValidationError{
			Field:   "members",
			Message: "members array is required and cannot be empty",
		})
	}

	for i, member := range req.Members {
		prefix := fmt.Sprintf("members[%d]", i)
		if member.UserID == "" {
			errors = append(errors, ValidationError{
				Field:   prefix + ".user_id",
				Message: "user_id is required",
			})
		}
		if member.Username == "" {
			errors = append(errors, ValidationError{
				Field:   prefix + ".username",
				Message: "username is required",
			})
		}
	}

	return errors
}

// ValidateSetUserActiveRequest валидирует SetUserActiveRequest
func ValidateSetUserActiveRequest(req dto.SetUserActiveRequest) []ValidationError {
	var errors []ValidationError

	if req.UserID == "" {
		errors = append(errors, ValidationError{
			Field:   "user_id",
			Message: "user_id is required",
		})
	}

	return errors
}

// ValidateCreatePRRequest валидирует CreatePRRequest
func ValidateCreatePRRequest(req dto.CreatePRRequest) []ValidationError {
	var errors []ValidationError

	if req.PullRequestID == "" {
		errors = append(errors, ValidationError{
			Field:   "pull_request_id",
			Message: "pull_request_id is required",
		})
	}

	if req.PullRequestName == "" {
		errors = append(errors, ValidationError{
			Field:   "pull_request_name",
			Message: "pull_request_name is required",
		})
	}

	if req.AuthorID == "" {
		errors = append(errors, ValidationError{
			Field:   "author_id",
			Message: "author_id is required",
		})
	}

	return errors
}

// ValidateMergePRRequest валидирует MergePRRequest
func ValidateMergePRRequest(req dto.MergePRRequest) []ValidationError {
	var errors []ValidationError

	if req.PullRequestID == "" {
		errors = append(errors, ValidationError{
			Field:   "pull_request_id",
			Message: "pull_request_id is required",
		})
	}

	return errors
}

// ValidateReassignReviewerRequest валидирует ReassignReviewerRequest
func ValidateReassignReviewerRequest(req dto.ReassignReviewerRequest) []ValidationError {
	var errors []ValidationError

	if req.PullRequestID == "" {
		errors = append(errors, ValidationError{
			Field:   "pull_request_id",
			Message: "pull_request_id is required",
		})
	}

	if req.OldUserID == "" {
		errors = append(errors, ValidationError{
			Field:   "old_user_id",
			Message: "old_user_id is required",
		})
	}

	return errors
}

// ValidateDeactivateTeamMembersRequest валидирует DeactivateTeamMembersRequest
func ValidateDeactivateTeamMembersRequest(req dto.DeactivateTeamMembersRequest) []ValidationError {
	var errors []ValidationError

	if strings.TrimSpace(req.TeamName) == "" {
		errors = append(errors, ValidationError{
			Field:   "team_name",
			Message: "team_name is required",
		})
	}

	return errors
}

// RespondValidationErrors отправляет ошибки валидации в формате API
func RespondValidationErrors(w http.ResponseWriter, errors []ValidationError) {
	if len(errors) == 0 {
		return
	}

	var messages []string
	for _, err := range errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}

	message := strings.Join(messages, "; ")
	presenter.RespondError(w, http.StatusBadRequest, presenter.ErrorCodeInvalidRequest, message)
}
