package entity

import (
	"fmt"
	"time"
)

// PRStatus представляет статус Pull Request
type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"

	MaxReviewersCount = 2
)

// PullRequest представляет Pull Request в доменной модели
type PullRequest struct {
	id                string
	name              string
	authorID          string
	status            PRStatus
	assignedReviewers []string
	createdAt         time.Time
	mergedAt          *time.Time // nullable заполняется при merge
}

// NewPullRequest создаёт новый Pull Request с валидацией
func NewPullRequest(id, name, authorID string) (*PullRequest, error) {
	normalizedID, err := validateAndNormalizeID(id)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidID, err)
	}

	normalizedName, err := validateAndNormalizePRName(name)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidPRName, err)
	}

	normalizedAuthorID, err := validateAndNormalizeID(authorID)
	if err != nil {
		return nil, fmt.Errorf("invalid author_id: %w: %w", ErrInvalidID, err)
	}

	now := time.Now().UTC()

	return &PullRequest{
		id:                normalizedID,
		name:              normalizedName,
		authorID:          normalizedAuthorID,
		status:            PRStatusOpen,
		assignedReviewers: []string{},
		createdAt:         now,
		mergedAt:          nil,
	}, nil
}

// NewPullRequestFromRepository восстанавливает PR из хранилища без валидации
func NewPullRequestFromRepository(
	id string,
	name string,
	authorID string,
	status PRStatus,
	assignedReviewers []string,
	createdAt time.Time,
	mergedAt *time.Time,
) *PullRequest {
	return &PullRequest{
		id:                id,
		name:              name,
		authorID:          authorID,
		status:            status,
		assignedReviewers: assignedReviewers,
		createdAt:         createdAt,
		mergedAt:          mergedAt,
	}
}

func (pr *PullRequest) ID() string {
	return pr.id
}

func (pr *PullRequest) Name() string {
	return pr.name
}

func (pr *PullRequest) AuthorID() string {
	return pr.authorID
}

func (pr *PullRequest) Status() PRStatus {
	return pr.status
}

func (pr *PullRequest) AssignedReviewers() []string {
	reviewers := make([]string, len(pr.assignedReviewers))
	copy(reviewers, pr.assignedReviewers)
	return reviewers
}

func (pr *PullRequest) CreatedAt() time.Time {
	return pr.createdAt
}

func (pr *PullRequest) MergedAt() *time.Time {
	return pr.mergedAt
}

// IsOpen возвращает true если PR в статусе OPEN
func (pr *PullRequest) IsOpen() bool {
	return pr.status == PRStatusOpen
}

// IsMerged возвращает true если PR смержен
func (pr *PullRequest) IsMerged() bool {
	return pr.status == PRStatusMerged
}

// AddReviewer добавляет ревьювера к PR
func (pr *PullRequest) AddReviewer(reviewerID string) error {
	if pr.IsMerged() {
		return ErrPRMerged
	}

	normalizedID, err := validateAndNormalizeID(reviewerID)
	if err != nil {
		return fmt.Errorf("invalid reviewer_id: %w: %w", ErrInvalidID, err)
	}

	if normalizedID == pr.authorID {
		return ErrAuthorCannotReview
	}

	for _, existingReviewer := range pr.assignedReviewers {
		if existingReviewer == normalizedID {
			return ErrReviewerAlreadyAssigned
		}
	}

	if len(pr.assignedReviewers) >= MaxReviewersCount {
		return ErrTooManyReviewers
	}

	pr.assignedReviewers = append(pr.assignedReviewers, normalizedID)

	return nil
}

// RemoveReviewer удаляет ревьювера из PR
func (pr *PullRequest) RemoveReviewer(reviewerID string) error {
	if pr.IsMerged() {
		return ErrPRMerged
	}

	normalizedID, err := validateAndNormalizeID(reviewerID)
	if err != nil {
		return fmt.Errorf("invalid reviewer_id: %w: %w", ErrInvalidID, err)
	}

	found := false
	newReviewers := make([]string, 0, len(pr.assignedReviewers)-1)
	for _, existingReviewer := range pr.assignedReviewers {
		if existingReviewer == normalizedID {
			found = true
			continue
		}
		newReviewers = append(newReviewers, existingReviewer)
	}

	if !found {
		return ErrReviewerNotAssigned
	}

	pr.assignedReviewers = newReviewers

	return nil
}

// ReplaceReviewer заменяет одного ревьювера на другого
func (pr *PullRequest) ReplaceReviewer(oldReviewerID, newReviewerID string) error {
	if pr.IsMerged() {
		return ErrPRMerged
	}

	normalizedOldID, err := validateAndNormalizeID(oldReviewerID)
	if err != nil {
		return fmt.Errorf("invalid old_reviewer_id: %w: %w", ErrInvalidID, err)
	}

	normalizedNewID, err := validateAndNormalizeID(newReviewerID)
	if err != nil {
		return fmt.Errorf("invalid new_reviewer_id: %w: %w", ErrInvalidID, err)
	}

	if normalizedNewID == pr.authorID {
		return ErrAuthorCannotReview
	}

	for _, existingReviewer := range pr.assignedReviewers {
		if existingReviewer == normalizedNewID {
			return ErrReviewerAlreadyAssigned
		}
	}

	found := false
	for i, existingReviewer := range pr.assignedReviewers {
		if existingReviewer == normalizedOldID {
			pr.assignedReviewers[i] = normalizedNewID
			found = true
			break
		}
	}

	if !found {
		return ErrReviewerNotAssigned
	}

	return nil
}

// Merge помечает PR как смерженный
func (pr *PullRequest) Merge() bool {
	if pr.IsMerged() {
		return false
	}

	pr.status = PRStatusMerged
	now := time.Now().UTC()
	pr.mergedAt = &now

	return true
}

// Equals сравнивает два PR по идентификатору
func (pr *PullRequest) Equals(other *PullRequest) bool {
	if other == nil {
		return false
	}
	return pr.id == other.id
}
