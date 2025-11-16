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

// TeamHandler обработчик для команд
type TeamHandler struct {
	teamUseCase TeamUseCase
}

// TeamUseCase интерфейс use case для команд (локальный для handler)
type TeamUseCase interface {
	CreateTeam(ctx context.Context, req dto.CreateTeamRequest) (*dto.TeamDTO, error)
	GetTeam(ctx context.Context, teamName string) (*dto.TeamDTO, error)
	DeactivateTeamMembers(ctx context.Context, teamName string) (*dto.TeamDTO, error)
}

// NewTeamHandler создает новый TeamHandler
func NewTeamHandler(teamUseCase TeamUseCase) *TeamHandler {
	return &TeamHandler{
		teamUseCase: teamUseCase,
	}
}

// CreateTeam обрабатывает POST /team/add
func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		presenter.RespondError(w, http.StatusBadRequest, presenter.ErrorCodeInvalidRequest, "invalid request body")
		return
	}

	if validationErrors := validator.ValidateCreateTeamRequest(req); len(validationErrors) > 0 {
		validator.RespondValidationErrors(w, validationErrors)
		return
	}

	team, err := h.teamUseCase.CreateTeam(r.Context(), req)
	if err != nil {
		statusCode, code, message := presenter.MapUseCaseError(err)
		presenter.RespondError(w, statusCode, code, message)
		return
	}

	presenter.RespondTeam(w, http.StatusCreated, team)
}

// GetTeam обрабатывает GET /team/get?team_name=
func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if strings.TrimSpace(teamName) == "" {
		presenter.RespondError(w, http.StatusBadRequest, presenter.ErrorCodeInvalidRequest, "team_name parameter is required")
		return
	}

	team, err := h.teamUseCase.GetTeam(r.Context(), teamName)
	if err != nil {
		statusCode, code, message := presenter.MapUseCaseError(err)
		presenter.RespondError(w, statusCode, code, message)
		return
	}

	presenter.RespondSuccess(w, http.StatusOK, team)
}

// DeactivateTeamMembers обрабатывает POST /team/deactivateMembers
func (h *TeamHandler) DeactivateTeamMembers(w http.ResponseWriter, r *http.Request) {
	var req dto.DeactivateTeamMembersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		presenter.RespondError(w, http.StatusBadRequest, presenter.ErrorCodeInvalidRequest, "invalid request body")
		return
	}

	if validationErrors := validator.ValidateDeactivateTeamMembersRequest(req); len(validationErrors) > 0 {
		validator.RespondValidationErrors(w, validationErrors)
		return
	}

	team, err := h.teamUseCase.DeactivateTeamMembers(r.Context(), req.TeamName)
	if err != nil {
		statusCode, code, message := presenter.MapUseCaseError(err)
		presenter.RespondError(w, statusCode, code, message)
		return
	}

	presenter.RespondTeam(w, http.StatusOK, team)
}

// RegisterRoutes регистрирует маршруты для команд
func (h *TeamHandler) RegisterRoutes(r chi.Router) {
	r.Post("/team/add", h.CreateTeam)
	r.Get("/team/get", h.GetTeam)
	r.Post("/team/deactivateMembers", h.DeactivateTeamMembers)
}
