package usecase

import (
	"context"
	"fmt"

	"github.com/exPriceD/pr-reviewer-service/internal/domain/logger"
	"github.com/exPriceD/pr-reviewer-service/internal/domain/repository"
	"github.com/exPriceD/pr-reviewer-service/internal/usecase/dto"
)

// StatisticsUseCase Use Case для получения статистики
type StatisticsUseCase struct {
	prRepo   repository.PullRequestRepository
	userRepo repository.UserRepository
	logger   logger.Logger
}

// NewStatisticsUseCase создает новый StatisticsUseCase
func NewStatisticsUseCase(
	prRepo repository.PullRequestRepository,
	userRepo repository.UserRepository,
	logger logger.Logger,
) *StatisticsUseCase {
	return &StatisticsUseCase{
		prRepo:   prRepo,
		userRepo: userRepo,
		logger:   logger,
	}
}

// GetStatistics возвращает статистику по назначениям
// Включает:
// - Статистику по PR (общее количество, открытые, смерженные)
// - Статистику по пользователям (количество назначений, активных назначений)
// Если teamName указан, возвращает статистику только для пользователей этой команды
// Если teamName пустой, возвращает статистику для всех пользователей
func (uc *StatisticsUseCase) GetStatistics(ctx context.Context, teamName string) (*dto.StatisticsDTO, error) {
	uc.logger.Info("Getting statistics", "team_name", teamName)

	total, open, merged, err := uc.prRepo.GetStats(ctx)
	if err != nil {
		uc.logger.Error("Failed to get PR stats", "error", err)
		return nil, fmt.Errorf("failed to get PR stats: %w", err)
	}

	result := &dto.StatisticsDTO{
		PRStats: dto.PRStatsDTO{
			Total:  total,
			Open:   open,
			Merged: merged,
		},
		UserStats: []dto.UserStatsDTO{},
	}

	if teamName != "" {
		users, err := uc.userRepo.FindByTeamName(ctx, teamName)
		if err != nil {
			uc.logger.Error("Failed to find team users", "error", err, "team_name", teamName)
			return nil, fmt.Errorf("failed to find team users: %w", err)
		}

		if len(users) == 0 {
			uc.logger.Info("Team has no users", "team_name", teamName)
			return result, nil
		}

		userIDs := make([]string, 0, len(users))
		for _, user := range users {
			userIDs = append(userIDs, user.ID())
		}

		totalReviews, err := uc.prRepo.CountReviewsByUserIDs(ctx, userIDs)
		if err != nil {
			uc.logger.Error("Failed to count total reviews", "error", err)
			return nil, fmt.Errorf("failed to count total reviews: %w", err)
		}

		activeReviews, err := uc.prRepo.CountActiveReviewsByUserIDs(ctx, userIDs)
		if err != nil {
			uc.logger.Error("Failed to count active reviews", "error", err)
			return nil, fmt.Errorf("failed to count active reviews: %w", err)
		}

		userStats := make([]dto.UserStatsDTO, 0, len(users))
		for _, user := range users {
			userID := user.ID()
			userStats = append(userStats, dto.UserStatsDTO{
				UserID:        userID,
				TotalReviews:  totalReviews[userID],
				ActiveReviews: activeReviews[userID],
			})
		}

		result.UserStats = userStats
	}

	uc.logger.Info("Statistics retrieved successfully", "pr_total", total, "pr_open", open, "pr_merged", merged)
	return result, nil
}
