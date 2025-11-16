package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/exPriceD/pr-reviewer-service/internal/delivery/http/presenter"
	"github.com/exPriceD/pr-reviewer-service/internal/usecase/dto"
)

// StatisticsHandler обработчик для статистики
type StatisticsHandler struct {
	statisticsUseCase StatisticsUseCase
}

// StatisticsUseCase интерфейс use case для статистики (локальный для handler)
type StatisticsUseCase interface {
	GetStatistics(ctx context.Context, teamName string) (*dto.StatisticsDTO, error)
}

// NewStatisticsHandler создает новый StatisticsHandler
func NewStatisticsHandler(statisticsUseCase StatisticsUseCase) *StatisticsHandler {
	return &StatisticsHandler{
		statisticsUseCase: statisticsUseCase,
	}
}

// GetStatistics обрабатывает GET /statistics?team_name=
// Если team_name указан, возвращает статистику для пользователей этой команды
// Если team_name не указан, возвращает только статистику по PR
func (h *StatisticsHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	teamName := strings.TrimSpace(r.URL.Query().Get("team_name"))

	stats, err := h.statisticsUseCase.GetStatistics(r.Context(), teamName)
	if err != nil {
		statusCode, code, message := presenter.MapUseCaseError(err)
		presenter.RespondError(w, statusCode, code, message)
		return
	}

	presenter.RespondStatistics(w, http.StatusOK, stats)
}

// RegisterRoutes регистрирует маршруты для статистики
func (h *StatisticsHandler) RegisterRoutes(r chi.Router) {
	r.Get("/statistics", h.GetStatistics)
}
