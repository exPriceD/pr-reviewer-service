package presenter

import (
	"errors"
	"net/http"

	"github.com/exPriceD/pr-reviewer-service/internal/usecase"
)

// MapUseCaseError преобразует ошибку use case в HTTP ошибку
func MapUseCaseError(err error) (int, string, string) {
	if errors.Is(err, usecase.ErrTeamAlreadyExists) {
		return http.StatusBadRequest, ErrorCodeTeamExists, "team_name already exists"
	}
	if errors.Is(err, usecase.ErrTeamNotFound) {
		return http.StatusNotFound, ErrorCodeNotFound, "team not found"
	}
	if errors.Is(err, usecase.ErrUserNotFound) {
		return http.StatusNotFound, ErrorCodeNotFound, "user not found"
	}
	if errors.Is(err, usecase.ErrPRAlreadyExists) {
		return http.StatusConflict, ErrorCodePRExists, "PR id already exists"
	}
	if errors.Is(err, usecase.ErrPRNotFound) {
		return http.StatusNotFound, ErrorCodeNotFound, "pull request not found"
	}
	if errors.Is(err, usecase.ErrPRAlreadyMerged) {
		return http.StatusConflict, ErrorCodePRMerged, "cannot reassign on merged PR"
	}
	if errors.Is(err, usecase.ErrReviewerNotAssigned) {
		return http.StatusConflict, ErrorCodeNotAssigned, "reviewer is not assigned to this PR"
	}
	if errors.Is(err, usecase.ErrNoActiveCandidates) {
		return http.StatusConflict, ErrorCodeNoCandidate, "no active replacement candidate in team"
	}
	return http.StatusInternalServerError, ErrorCodeInternalError, "internal server error"
}
