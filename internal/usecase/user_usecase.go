package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/exPriceD/pr-reviewer-service/internal/domain/logger"
	"github.com/exPriceD/pr-reviewer-service/internal/domain/repository"
	"github.com/exPriceD/pr-reviewer-service/internal/usecase/dto"
)

// UserUseCase Use Case для работы с пользователями
type UserUseCase struct {
	userRepo repository.UserRepository
	prRepo   repository.PullRequestRepository
	logger   logger.Logger
}

// NewUserUseCase создает новый UserUseCase
func NewUserUseCase(
	userRepo repository.UserRepository,
	prRepo repository.PullRequestRepository,
	logger logger.Logger,
) *UserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
		prRepo:   prRepo,
		logger:   logger,
	}
}

// SetUserActive устанавливает флаг активности пользователя
// POST /users/setIsActive
func (uc *UserUseCase) SetUserActive(ctx context.Context, req dto.SetUserActiveRequest) (*dto.UserDTO, error) {
	uc.logger.Info("Setting user active status", "user_id", req.UserID, "is_active", req.IsActive)

	user, err := uc.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		uc.logger.Error("Failed to find user", "error", err, "user_id", req.UserID)
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if req.IsActive && !user.IsActive() {
		user.Activate()
	} else if !req.IsActive && user.IsActive() {
		user.Deactivate()
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		uc.logger.Error("Failed to update user", "error", err, "user_id", req.UserID)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	uc.logger.Info("User active status updated", "user_id", req.UserID, "is_active", req.IsActive)
	result := dto.ToUserDTO(user)
	return &result, nil
}

// GetUserReviews получает PR'ы где пользователь назначен ревьювером
// GET /users/getReview?user_id=
func (uc *UserUseCase) GetUserReviews(ctx context.Context, userID string) ([]dto.PullRequestShortDTO, error) {
	uc.logger.Info("Getting user reviews", "user_id", userID)

	exists, err := uc.userRepo.Exists(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to check user existence", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	prs, err := uc.prRepo.FindByReviewerID(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to find user PRs", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to find user PRs: %w", err)
	}

	uc.logger.Info("User reviews retrieved", "user_id", userID, "prs_count", len(prs))
	return dto.ToPullRequestShortDTOs(prs), nil
}
