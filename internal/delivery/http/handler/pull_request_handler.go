package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/exPriceD/pr-reviewer-service/internal/delivery/http/presenter"
	"github.com/exPriceD/pr-reviewer-service/internal/delivery/http/validator"
	"github.com/exPriceD/pr-reviewer-service/internal/usecase/dto"
)

// PullRequestHandler обработчик для Pull Requests
type PullRequestHandler struct {
	prUseCase PullRequestUseCase
}

// PullRequestUseCase интерфейс use case для Pull Requests (локальный для handler)
type PullRequestUseCase interface {
	CreatePR(ctx context.Context, req dto.CreatePRRequest) (*dto.PullRequestDTO, error)
	MergePR(ctx context.Context, prID string) (*dto.PullRequestDTO, error)
	ReassignReviewer(ctx context.Context, req dto.ReassignReviewerRequest) (*dto.PullRequestDTO, string, error)
}

// NewPullRequestHandler создает новый PullRequestHandler
func NewPullRequestHandler(prUseCase PullRequestUseCase) *PullRequestHandler {
	return &PullRequestHandler{
		prUseCase: prUseCase,
	}
}

// CreatePR обрабатывает POST /pullRequest/create
func (h *PullRequestHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req dto.CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		presenter.RespondError(w, http.StatusBadRequest, presenter.ErrorCodeInvalidRequest, "invalid request body")
		return
	}

	if validationErrors := validator.ValidateCreatePRRequest(req); len(validationErrors) > 0 {
		validator.RespondValidationErrors(w, validationErrors)
		return
	}

	pr, err := h.prUseCase.CreatePR(r.Context(), req)
	if err != nil {
		statusCode, code, message := presenter.MapUseCaseError(err)
		presenter.RespondError(w, statusCode, code, message)
		return
	}

	presenter.RespondPullRequest(w, http.StatusCreated, pr)
}

// MergePR обрабатывает POST /pullRequest/merge
func (h *PullRequestHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	var req dto.MergePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		presenter.RespondError(w, http.StatusBadRequest, presenter.ErrorCodeInvalidRequest, "invalid request body")
		return
	}

	if validationErrors := validator.ValidateMergePRRequest(req); len(validationErrors) > 0 {
		validator.RespondValidationErrors(w, validationErrors)
		return
	}

	pr, err := h.prUseCase.MergePR(r.Context(), req.PullRequestID)
	if err != nil {
		statusCode, code, message := presenter.MapUseCaseError(err)
		presenter.RespondError(w, statusCode, code, message)
		return
	}

	presenter.RespondPullRequest(w, http.StatusOK, pr)
}

// ReassignReviewer обрабатывает POST /pullRequest/reassign
func (h *PullRequestHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	var req dto.ReassignReviewerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		presenter.RespondError(w, http.StatusBadRequest, presenter.ErrorCodeInvalidRequest, "invalid request body")
		return
	}

	if validationErrors := validator.ValidateReassignReviewerRequest(req); len(validationErrors) > 0 {
		validator.RespondValidationErrors(w, validationErrors)
		return
	}

	pr, replacedBy, err := h.prUseCase.ReassignReviewer(r.Context(), req)
	if err != nil {
		statusCode, code, message := presenter.MapUseCaseError(err)
		presenter.RespondError(w, statusCode, code, message)
		return
	}

	presenter.RespondPullRequestReassign(w, http.StatusOK, pr, replacedBy)
}

// RegisterRoutes регистрирует маршруты для Pull Requests
func (h *PullRequestHandler) RegisterRoutes(r chi.Router) {
	r.Post("/pullRequest/create", h.CreatePR)
	r.Post("/pullRequest/merge", h.MergePR)
	r.Post("/pullRequest/reassign", h.ReassignReviewer)
}
