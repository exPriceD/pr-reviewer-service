package pull_request

import (
	"database/sql"
	"time"

	"github.com/exPriceD/pr-reviewer-service/internal/domain/entity"
)

func ToEntity(m *Model, reviewers []string) *entity.PullRequest {
	var mergedAtPtr *time.Time
	if m.MergedAt.Valid {
		mergedAtPtr = &m.MergedAt.Time
	}

	return entity.NewPullRequestFromRepository(
		m.ID,
		m.Name,
		m.AuthorID,
		entity.PRStatus(m.Status),
		reviewers,
		m.CreatedAt,
		mergedAtPtr,
	)
}

func FromEntity(pr *entity.PullRequest) *Model {
	var mergedAt sql.NullTime
	if pr.MergedAt() != nil {
		mergedAt = sql.NullTime{
			Time:  *pr.MergedAt(),
			Valid: true,
		}
	}

	return &Model{
		ID:        pr.ID(),
		Name:      pr.Name(),
		AuthorID:  pr.AuthorID(),
		Status:    string(pr.Status()),
		CreatedAt: pr.CreatedAt(),
		MergedAt:  mergedAt,
	}
}
