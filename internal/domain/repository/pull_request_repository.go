package repository

import (
	"context"

	"github.com/exPriceD/pr-reviewer-service/internal/domain/entity"
)

// PullRequestRepository интерфейс для работы с Pull Request
type PullRequestRepository interface {
	Create(ctx context.Context, pr *entity.PullRequest) error
	FindByID(ctx context.Context, id string) (*entity.PullRequest, error)
	FindByIDForUpdate(ctx context.Context, id string) (*entity.PullRequest, error)
	FindByReviewerID(ctx context.Context, reviewerID string) ([]*entity.PullRequest, error)
	FindByAuthorID(ctx context.Context, authorID string) ([]*entity.PullRequest, error)
	Update(ctx context.Context, pr *entity.PullRequest) error
	ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
	MergePR(ctx context.Context, prID string) error
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)
	CountActiveReviewsByUserIDs(ctx context.Context, userIDs []string) (map[string]int, error)
	GetStats(ctx context.Context) (total, open, merged int, err error)
	CountReviewsByUserIDs(ctx context.Context, userIDs []string) (map[string]int, error)
}
