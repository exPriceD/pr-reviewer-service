package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/exPriceD/pr-reviewer-service/internal/delivery/http/presenter"
	"github.com/exPriceD/pr-reviewer-service/internal/delivery/http/validator"
	"github.com/exPriceD/pr-reviewer-service/internal/usecase/dto"
)

// UserHandler обработчик для пользователей
type UserHandler struct {
	userUseCase UserUseCase
}

// UserUseCase интерфейс use case для пользователей (локальный для handler)
type UserUseCase interface {
	SetUserActive(ctx context.Context, req dto.SetUserActiveRequest) (*dto.UserDTO, error)
	GetUserReviews(ctx context.Context, userID string) ([]dto.PullRequestShortDTO, error)
}

// NewUserHandler создает новый UserHandler
func NewUserHandler(userUseCase UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

// SetUserActive обрабатывает POST /users/setIsActive
func (h *UserHandler) SetUserActive(w http.ResponseWriter, r *http.Request) {
	var req dto.SetUserActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		presenter.RespondError(w, http.StatusBadRequest, presenter.ErrorCodeInvalidRequest, "invalid request body")
		return
	}

	if validationErrors := validator.ValidateSetUserActiveRequest(req); len(validationErrors) > 0 {
		validator.RespondValidationErrors(w, validationErrors)
		return
	}

	user, err := h.userUseCase.SetUserActive(r.Context(), req)
	if err != nil {
		statusCode, code, message := presenter.MapUseCaseError(err)
		presenter.RespondError(w, statusCode, code, message)
		return
	}

	presenter.RespondUser(w, http.StatusOK, user)
}

// GetUserReviews обрабатывает GET /users/getReview?user_id=
func (h *UserHandler) GetUserReviews(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" {
		presenter.RespondError(w, http.StatusBadRequest, presenter.ErrorCodeInvalidRequest, "user_id parameter is required")
		return
	}

	prs, err := h.userUseCase.GetUserReviews(r.Context(), userID)
	if err != nil {
		statusCode, code, message := presenter.MapUseCaseError(err)
		presenter.RespondError(w, statusCode, code, message)
		return
	}

	presenter.RespondUserReviews(w, http.StatusOK, userID, prs)
}

// RegisterRoutes регистрирует маршруты для пользователей
func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.Post("/users/setIsActive", h.SetUserActive)
	r.Get("/users/getReview", h.GetUserReviews)
}
