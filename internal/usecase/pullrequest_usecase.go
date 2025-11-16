package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/exPriceD/pr-reviewer-service/internal/domain/entity"
	"github.com/exPriceD/pr-reviewer-service/internal/domain/logger"
	"github.com/exPriceD/pr-reviewer-service/internal/domain/repository"
	"github.com/exPriceD/pr-reviewer-service/internal/domain/transaction"
	"github.com/exPriceD/pr-reviewer-service/internal/usecase/dto"
)

// PullRequestUseCase Use Case для работы с Pull Requests
type PullRequestUseCase struct {
	txManager        transaction.Manager
	prRepo           repository.PullRequestRepository
	userRepo         repository.UserRepository
	reviewerSelector *ReviewerSelector
	logger           logger.Logger
}

// NewPullRequestUseCase создает новый PullRequestUseCase
func NewPullRequestUseCase(
	txManager transaction.Manager,
	prRepo repository.PullRequestRepository,
	userRepo repository.UserRepository,
	reviewerSelector *ReviewerSelector,
	logger logger.Logger,
) *PullRequestUseCase {
	return &PullRequestUseCase{
		txManager:        txManager,
		prRepo:           prRepo,
		userRepo:         userRepo,
		reviewerSelector: reviewerSelector,
		logger:           logger,
	}
}

// CreatePR создает PR и автоматически назначает до MaxReviewersCount ревьюеров из команды автора
// POST /pullRequest/create
func (uc *PullRequestUseCase) CreatePR(ctx context.Context, req dto.CreatePRRequest) (*dto.PullRequestDTO, error) {
	uc.logger.Info("Creating PR", "pr_id", req.PullRequestID, "author_id", req.AuthorID)

	exists, err := uc.prRepo.Exists(ctx, req.PullRequestID)
	if err != nil {
		uc.logger.Error("Failed to check PR existence", "error", err)
		return nil, fmt.Errorf("failed to check PR existence: %w", err)
	}
	if exists {
		return nil, ErrPRAlreadyExists
	}

	author, err := uc.userRepo.FindByID(ctx, req.AuthorID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		uc.logger.Error("Failed to find author", "error", err, "author_id", req.AuthorID)
		return nil, fmt.Errorf("failed to find author: %w", err)
	}

	var pr *entity.PullRequest

	err = uc.txManager.Do(ctx, func(ctx context.Context) error {
		pr, err = entity.NewPullRequest(req.PullRequestID, req.PullRequestName, req.AuthorID)
		if err != nil {
			return fmt.Errorf("failed to create PR entity: %w", err)
		}

		reviewers, err := uc.reviewerSelector.SelectReviewers(ctx, author.TeamName(), req.AuthorID)
		if err != nil {
			return fmt.Errorf("failed to select reviewers: %w", err)
		}

		for _, reviewerID := range reviewers {
			if err := pr.AddReviewer(reviewerID); err != nil {
				return fmt.Errorf("failed to add reviewer %s: %w", reviewerID, err)
			}
		}

		if err := uc.prRepo.Create(ctx, pr); err != nil {
			return fmt.Errorf("failed to save PR: %w", err)
		}
		return nil
	})
	if err != nil {
		uc.logger.Error("Failed to create PR", "error", err, "pr_id", req.PullRequestID)
		return nil, err
	}

	uc.logger.Info("PR created successfully",
		"pr_id", req.PullRequestID,
		"reviewers_count", len(pr.AssignedReviewers()),
		"reviewers", pr.AssignedReviewers(),
	)
	result := dto.ToPullRequestDTO(pr)
	return &result, nil
}

// MergePR помечает PR как MERGED (идемпотентная операция)
// POST /pullRequest/merge
func (uc *PullRequestUseCase) MergePR(ctx context.Context, prID string) (*dto.PullRequestDTO, error) {
	uc.logger.Info("Merging PR", "pr_id", prID)

	err := uc.prRepo.MergePR(ctx, prID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			pr, getErr := uc.prRepo.FindByID(ctx, prID)
			if getErr != nil {
				if errors.Is(getErr, repository.ErrNotFound) {
					return nil, ErrPRNotFound
				}
				uc.logger.Error("Failed to find PR after merge attempt", "error", getErr, "pr_id", prID)
				return nil, fmt.Errorf("failed to find PR: %w", getErr)
			}

			if pr.IsMerged() {
				uc.logger.Info("PR already merged (idempotent operation)", "pr_id", prID)
				result := dto.ToPullRequestDTO(pr)
				return &result, nil
			}

			uc.logger.Error("Unexpected state: PR exists but CTE failed", "pr_id", prID)
			return nil, fmt.Errorf("failed to merge PR: unexpected state")
		}

		uc.logger.Error("Failed to merge PR", "error", err, "pr_id", prID)
		return nil, fmt.Errorf("failed to merge PR: %w", err)
	}

	pr, err := uc.prRepo.FindByID(ctx, prID)
	if err != nil {
		uc.logger.Error("Failed to find PR after merge", "error", err, "pr_id", prID)
		return nil, fmt.Errorf("failed to find PR after merge: %w", err)
	}

	uc.logger.Info("PR merged successfully", "pr_id", prID)
	result := dto.ToPullRequestDTO(pr)
	return &result, nil
}

// ReassignReviewer переназначает ревьювера на случайного активного из команды заменяемого
// POST /pullRequest/reassign
func (uc *PullRequestUseCase) ReassignReviewer(ctx context.Context, req dto.ReassignReviewerRequest) (*dto.PullRequestDTO, string, error) {
	uc.logger.Info("Reassigning reviewer",
		"pr_id", req.PullRequestID,
		"old_user_id", req.OldUserID,
	)

	var pr *entity.PullRequest
	var newReviewerID string

	err := uc.txManager.Do(ctx, func(ctx context.Context) error {
		var err error
		pr, err = uc.prRepo.FindByIDForUpdate(ctx, req.PullRequestID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrPRNotFound
			}
			return fmt.Errorf("failed to find PR for update: %w", err)
		}

		if pr.IsMerged() {
			return ErrPRAlreadyMerged
		}

		isAssigned := false
		for _, reviewerID := range pr.AssignedReviewers() {
			if reviewerID == req.OldUserID {
				isAssigned = true
				break
			}
		}
		if !isAssigned {
			return ErrReviewerNotAssigned
		}

		newReviewerID, err = uc.reviewerSelector.SelectReplacement(
			ctx,
			req.OldUserID,
			pr.AuthorID(),
			pr.AssignedReviewers(),
		)
		if err != nil {
			return err
		}

		if err := pr.ReplaceReviewer(req.OldUserID, newReviewerID); err != nil {
			return fmt.Errorf("failed to replace reviewer in entity: %w", err)
		}

		if err := uc.prRepo.ReplaceReviewer(ctx, pr.ID(), req.OldUserID, newReviewerID); err != nil {
			return fmt.Errorf("failed to replace reviewer in database: %w", err)
		}

		return nil
	})
	if err != nil {
		uc.logger.Error("Failed to reassign reviewer",
			"error", err,
			"pr_id", req.PullRequestID,
			"old_reviewer_id", req.OldUserID,
		)
		return nil, "", err
	}

	uc.logger.Info("Reviewer reassigned successfully",
		"pr_id", req.PullRequestID,
		"old_reviewer_id", req.OldUserID,
		"new_reviewer_id", newReviewerID,
	)
	result := dto.ToPullRequestDTO(pr)
	return &result, newReviewerID, nil
}
