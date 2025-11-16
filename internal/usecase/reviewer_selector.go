package usecase

import (
	"context"
	"fmt"
	"sort"

	"github.com/exPriceD/pr-reviewer-service/internal/domain/entity"
	"github.com/exPriceD/pr-reviewer-service/internal/domain/repository"
)

// ReviewerSelector сервис для выбора ревьюеров по алгоритму Round-Robin
type ReviewerSelector struct {
	userRepo repository.UserRepository
	prRepo   repository.PullRequestRepository
}

// NewReviewerSelector создает новый ReviewerSelector
func NewReviewerSelector(userRepo repository.UserRepository, prRepo repository.PullRequestRepository) *ReviewerSelector {
	return &ReviewerSelector{
		userRepo: userRepo,
		prRepo:   prRepo,
	}
}

// SelectReviewers выбирает до MaxReviewersCount активных ревьюеров из команды по алгоритму Round-Robin
// выбираются те, у кого меньше всего активных (OPEN) PR на ревью
func (s *ReviewerSelector) SelectReviewers(ctx context.Context, teamName, authorID string) ([]string, error) {
	users, err := s.userRepo.FindActiveByTeamName(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to find active team members: %w", err)
	}

	var candidateIDs []string
	for _, user := range users {
		if user.ID() != authorID {
			candidateIDs = append(candidateIDs, user.ID())
		}
	}

	if len(candidateIDs) == 0 {
		return []string{}, nil
	}

	reviewCounts, err := s.prRepo.CountActiveReviewsByUserIDs(ctx, candidateIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get review counts: %w", err)
	}

	return s.selectByRoundRobin(candidateIDs, reviewCounts, entity.MaxReviewersCount), nil
}

// SelectReplacement выбирает замену для ревьювера из его команды
func (s *ReviewerSelector) SelectReplacement(ctx context.Context, oldReviewerID, authorID string, assignedReviewers []string) (string, error) {
	oldReviewer, err := s.userRepo.FindByID(ctx, oldReviewerID)
	if err != nil {
		return "", fmt.Errorf("failed to find old reviewer: %w", err)
	}
	if oldReviewer == nil {
		return "", fmt.Errorf("old reviewer is nil")
	}

	users, err := s.userRepo.FindActiveByTeamName(ctx, oldReviewer.TeamName())
	if err != nil {
		return "", fmt.Errorf("failed to find active team members: %w", err)
	}

	exclude := make(map[string]bool)
	exclude[authorID] = true
	exclude[oldReviewerID] = true
	for _, id := range assignedReviewers {
		if id != oldReviewerID {
			exclude[id] = true
		}
	}

	var candidateIDs []string
	for _, user := range users {
		if !exclude[user.ID()] {
			candidateIDs = append(candidateIDs, user.ID())
		}
	}

	if len(candidateIDs) == 0 {
		return "", ErrNoActiveCandidates
	}

	reviewCounts, err := s.prRepo.CountActiveReviewsByUserIDs(ctx, candidateIDs)
	if err != nil {
		return "", fmt.Errorf("failed to get review counts: %w", err)
	}

	selected := s.selectByRoundRobin(candidateIDs, reviewCounts, 1)
	if len(selected) == 0 {
		return "", ErrNoActiveCandidates
	}

	return selected[0], nil
}

// selectByRoundRobin выбирает до maxCount пользователей с наименьшей загрузкой
func (s *ReviewerSelector) selectByRoundRobin(candidateIDs []string, reviewCounts map[string]int, maxCount int) []string {
	if len(candidateIDs) == 0 {
		return []string{}
	}

	if len(candidateIDs) <= maxCount {
		return candidateIDs
	}

	type candidate struct {
		userID string
		count  int
	}

	candidates := make([]candidate, 0, len(candidateIDs))
	for _, userID := range candidateIDs {
		candidates = append(candidates, candidate{
			userID: userID,
			count:  reviewCounts[userID],
		})
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].count == candidates[j].count {
			return candidates[i].userID < candidates[j].userID
		}
		return candidates[i].count < candidates[j].count
	})

	result := make([]string, maxCount)
	for i := 0; i < maxCount; i++ {
		result[i] = candidates[i].userID
	}

	return result
}
