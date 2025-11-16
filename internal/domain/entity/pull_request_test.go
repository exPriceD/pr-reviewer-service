package entity

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// TestNewPullRequest проверяет создание нового PR
func TestNewPullRequest(t *testing.T) {
	tests := []struct {
		name        string
		prID        string
		prName      string
		authorID    string
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "valid PR",
			prID:     "pr-1001",
			prName:   "Add user authentication",
			authorID: "user-123",
			wantErr:  false,
		},
		{
			name:     "valid PR with underscores",
			prID:     "pr_1002",
			prName:   "Fix_bug_in_payment",
			authorID: "user_456",
			wantErr:  false,
		},
		{
			name:     "valid PR with spaces trimmed",
			prID:     "  pr-1003  ",
			prName:   "  Refactor API  ",
			authorID: "  user-789  ",
			wantErr:  false,
		},
		{
			name:        "empty PR ID",
			prID:        "",
			prName:      "Some PR",
			authorID:    "user-1",
			wantErr:     true,
			expectedErr: ErrInvalidID,
		},
		{
			name:        "empty PR name",
			prID:        "pr-1",
			prName:      "",
			authorID:    "user-1",
			wantErr:     true,
			expectedErr: ErrInvalidPRName,
		},
		{
			name:        "empty author ID",
			prID:        "pr-1",
			prName:      "Some PR",
			authorID:    "",
			wantErr:     true,
			expectedErr: ErrInvalidID,
		},
		{
			name:        "PR name too long",
			prID:        "pr-1",
			prName:      strings.Repeat("a", 201),
			authorID:    "user-1",
			wantErr:     true,
			expectedErr: ErrInvalidPRName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr, err := NewPullRequest(tt.prID, tt.prName, tt.authorID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewPullRequest() expected error, got nil")
					return
				}
				if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
					t.Errorf("NewPullRequest() error = %v, want %v", err, tt.expectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("NewPullRequest() unexpected error: %v", err)
				return
			}

			if pr.ID() != strings.TrimSpace(tt.prID) {
				t.Errorf("ID = %v, want %v", pr.ID(), strings.TrimSpace(tt.prID))
			}
			if pr.Name() != strings.TrimSpace(tt.prName) {
				t.Errorf("Name = %v, want %v", pr.Name(), strings.TrimSpace(tt.prName))
			}
			if pr.AuthorID() != strings.TrimSpace(tt.authorID) {
				t.Errorf("AuthorID = %v, want %v", pr.AuthorID(), strings.TrimSpace(tt.authorID))
			}

			if pr.Status() != PRStatusOpen {
				t.Errorf("Status = %v, want %v", pr.Status(), PRStatusOpen)
			}
			if len(pr.AssignedReviewers()) != 0 {
				t.Errorf("AssignedReviewers length = %d, want 0", len(pr.AssignedReviewers()))
			}
			if !pr.IsOpen() {
				t.Errorf("IsOpen() = false, want true")
			}
			if pr.IsMerged() {
				t.Errorf("IsMerged() = true, want false")
			}
			if pr.MergedAt() != nil {
				t.Errorf("MergedAt = %v, want nil", pr.MergedAt())
			}
		})
	}
}

// TestPullRequestAddReviewer проверяет добавление ревьюверов
func TestPullRequestAddReviewer(t *testing.T) {
	pr, _ := NewPullRequest("pr-1", "Test PR", "author-1")

	err := pr.AddReviewer("reviewer-1")
	if err != nil {
		t.Fatalf("AddReviewer() failed: %v", err)
	}
	if len(pr.AssignedReviewers()) != 1 {
		t.Errorf("AssignedReviewers length = %d, want 1", len(pr.AssignedReviewers()))
	}
	if pr.AssignedReviewers()[0] != "reviewer-1" {
		t.Errorf("First reviewer = %v, want reviewer-1", pr.AssignedReviewers()[0])
	}

	err = pr.AddReviewer("reviewer-2")
	if err != nil {
		t.Fatalf("AddReviewer() failed: %v", err)
	}
	if len(pr.AssignedReviewers()) != MaxReviewersCount {
		t.Errorf("AssignedReviewers length = %d, want %d", len(pr.AssignedReviewers()), MaxReviewersCount)
	}

	err = pr.AddReviewer("reviewer-3")
	if !errors.Is(err, ErrTooManyReviewers) {
		t.Errorf("AddReviewer() error = %v, want ErrTooManyReviewers", err)
	}
}

// TestPullRequestAddReviewerValidation проверяет валидацию при добавлении
func TestPullRequestAddReviewerValidation(t *testing.T) {
	pr, _ := NewPullRequest("pr-1", "Test PR", "author-1")

	err := pr.AddReviewer("author-1")
	if !errors.Is(err, ErrAuthorCannotReview) {
		t.Errorf("AddReviewer(author) error = %v, want ErrAuthorCannotReview", err)
	}

	//nolint:errcheck
	_ = pr.AddReviewer("reviewer-1")

	err = pr.AddReviewer("reviewer-1")
	if !errors.Is(err, ErrReviewerAlreadyAssigned) {
		t.Errorf("AddReviewer(duplicate) error = %v, want ErrReviewerAlreadyAssigned", err)
	}

	err = pr.AddReviewer("")
	if !errors.Is(err, ErrInvalidID) {
		t.Errorf("AddReviewer(empty) error = %v, want ErrInvalidID", err)
	}
}

// TestPullRequestRemoveReviewer проверяет удаление ревьюверов
func TestPullRequestRemoveReviewer(t *testing.T) {
	pr, _ := NewPullRequest("pr-1", "Test PR", "author-1")
	//nolint:errcheck
	_ = pr.AddReviewer("reviewer-1")
	//nolint:errcheck
	_ = pr.AddReviewer("reviewer-2")

	err := pr.RemoveReviewer("reviewer-1")
	if err != nil {
		t.Fatalf("RemoveReviewer() failed: %v", err)
	}
	if len(pr.AssignedReviewers()) != 1 {
		t.Errorf("AssignedReviewers length = %d, want 1", len(pr.AssignedReviewers()))
	}
	if pr.AssignedReviewers()[0] != "reviewer-2" {
		t.Errorf("Remaining reviewer = %v, want reviewer-2", pr.AssignedReviewers()[0])
	}

	err = pr.RemoveReviewer("reviewer-1")
	if !errors.Is(err, ErrReviewerNotAssigned) {
		t.Errorf("RemoveReviewer(not assigned) error = %v, want ErrReviewerNotAssigned", err)
	}

	err = pr.RemoveReviewer("reviewer-2")
	if err != nil {
		t.Fatalf("RemoveReviewer() failed: %v", err)
	}
	if len(pr.AssignedReviewers()) != 0 {
		t.Errorf("AssignedReviewers length = %d, want 0", len(pr.AssignedReviewers()))
	}
}

// TestPullRequestReplaceReviewer проверяет замену ревьювера
func TestPullRequestReplaceReviewer(t *testing.T) {
	pr, _ := NewPullRequest("pr-1", "Test PR", "author-1")
	_ = pr.AddReviewer("reviewer-1")
	_ = pr.AddReviewer("reviewer-2")

	err := pr.ReplaceReviewer("reviewer-1", "reviewer-3")
	if err != nil {
		t.Fatalf("ReplaceReviewer() failed: %v", err)
	}

	reviewers := pr.AssignedReviewers()
	if len(reviewers) != MaxReviewersCount {
		t.Errorf("AssignedReviewers length = %d, want %d", len(reviewers), MaxReviewersCount)
	}
	if reviewers[0] != "reviewer-3" {
		t.Errorf("First reviewer = %v, want reviewer-3", reviewers[0])
	}
	if reviewers[1] != "reviewer-2" {
		t.Errorf("Second reviewer = %v, want reviewer-2", reviewers[1])
	}

	err = pr.ReplaceReviewer("reviewer-999", "reviewer-4")
	if !errors.Is(err, ErrReviewerNotAssigned) {
		t.Errorf("ReplaceReviewer(not assigned) error = %v, want ErrReviewerNotAssigned", err)
	}

	err = pr.ReplaceReviewer("reviewer-3", "reviewer-2")
	if !errors.Is(err, ErrReviewerAlreadyAssigned) {
		t.Errorf("ReplaceReviewer(already assigned) error = %v, want ErrReviewerAlreadyAssigned", err)
	}

	err = pr.ReplaceReviewer("reviewer-3", "author-1")
	if !errors.Is(err, ErrAuthorCannotReview) {
		t.Errorf("ReplaceReviewer(author) error = %v, want ErrAuthorCannotReview", err)
	}
}

// TestPullRequestMerge проверяет merge PR
func TestPullRequestMerge(t *testing.T) {
	pr, _ := NewPullRequest("pr-1", "Test PR", "author-1")

	if pr.IsMerged() {
		t.Errorf("IsMerged() = true, want false initially")
	}
	if pr.MergedAt() != nil {
		t.Errorf("MergedAt() = %v, want nil initially", pr.MergedAt())
	}

	changed := pr.Merge()
	if !changed {
		t.Errorf("Merge() = false, want true on first call")
	}
	if !pr.IsMerged() {
		t.Errorf("IsMerged() = false, want true after merge")
	}
	if pr.Status() != PRStatusMerged {
		t.Errorf("Status = %v, want %v after merge", pr.Status(), PRStatusMerged)
	}
	if pr.MergedAt() == nil {
		t.Errorf("MergedAt() = nil, want non-nil after merge")
	}

	changed = pr.Merge()
	if changed {
		t.Errorf("Merge() = true, want false on second call (idempotent)")
	}
}

// TestPullRequestOperationsAfterMerge проверяет, что после merge нельзя менять ревьюверов
func TestPullRequestOperationsAfterMerge(t *testing.T) {
	pr, _ := NewPullRequest("pr-1", "Test PR", "author-1")
	_ = pr.AddReviewer("reviewer-1")
	pr.Merge()

	err := pr.AddReviewer("reviewer-2")
	if !errors.Is(err, ErrPRMerged) {
		t.Errorf("AddReviewer() after merge error = %v, want ErrPRMerged", err)
	}

	err = pr.RemoveReviewer("reviewer-1")
	if !errors.Is(err, ErrPRMerged) {
		t.Errorf("RemoveReviewer() after merge error = %v, want ErrPRMerged", err)
	}

	err = pr.ReplaceReviewer("reviewer-1", "reviewer-2")
	if !errors.Is(err, ErrPRMerged) {
		t.Errorf("ReplaceReviewer() after merge error = %v, want ErrPRMerged", err)
	}
}

// TestPullRequestAssignedReviewersImmutability проверяет защиту от изменения списка
func TestPullRequestAssignedReviewersImmutability(t *testing.T) {
	pr, _ := NewPullRequest("pr-1", "Test PR", "author-1")
	_ = pr.AddReviewer("reviewer-1")
	_ = pr.AddReviewer("reviewer-2")

	reviewers := pr.AssignedReviewers()

	reviewers[0] = "hacker"
	_ = append(reviewers, "another-hacker")

	actualReviewers := pr.AssignedReviewers()
	if actualReviewers[0] == "hacker" {
		t.Errorf("Internal state was modified! AssignedReviewers() should return a copy")
	}
	if len(actualReviewers) != MaxReviewersCount {
		t.Errorf("Internal state was modified! Length changed from %d to %d", MaxReviewersCount, len(actualReviewers))
	}
}

// TestPullRequestEquals проверяет сравнение PR
func TestPullRequestEquals(t *testing.T) {
	pr1, _ := NewPullRequest("pr-1", "Test PR 1", "author-1")
	pr2, _ := NewPullRequest("pr-1", "Different Name", "author-2")
	pr3, _ := NewPullRequest("pr-2", "Test PR 2", "author-1")

	if !pr1.Equals(pr2) {
		t.Errorf("pr1.Equals(pr2) = false, want true (same ID)")
	}

	if pr1.Equals(pr3) {
		t.Errorf("pr1.Equals(pr3) = true, want false (different ID)")
	}

	if pr1.Equals(nil) {
		t.Errorf("pr1.Equals(nil) = true, want false")
	}

	//nolint:gocritic // Testing that object equals itself is intentional
	if !pr1.Equals(pr1) {
		t.Errorf("pr1.Equals(pr1) = false, want true (same object)")
	}
}

// TestPullRequestLifecycle проверяет полный жизненный цикл PR
func TestPullRequestLifecycle(t *testing.T) {
	pr, err := NewPullRequest("pr-123", "Implement feature X", "john-doe")
	if err != nil {
		t.Fatalf("Failed to create PR: %v", err)
	}

	if err := pr.AddReviewer("alice"); err != nil {
		t.Fatalf("Failed to add first reviewer: %v", err)
	}
	if err := pr.AddReviewer("bob"); err != nil {
		t.Fatalf("Failed to add second reviewer: %v", err)
	}

	if err := pr.ReplaceReviewer("alice", "charlie"); err != nil {
		t.Fatalf("Failed to replace reviewer: %v", err)
	}

	if !pr.Merge() {
		t.Errorf("Merge() = false, want true")
	}

	if !pr.IsMerged() {
		t.Errorf("IsMerged() = false, want true")
	}
	if pr.IsOpen() {
		t.Errorf("IsOpen() = true, want false")
	}

	reviewers := pr.AssignedReviewers()
	if len(reviewers) != MaxReviewersCount {
		t.Errorf("Final reviewers count = %d, want %d", len(reviewers), MaxReviewersCount)
	}

	if err := pr.AddReviewer("dan"); !errors.Is(err, ErrPRMerged) {
		t.Errorf("Operations after merge should be forbidden")
	}
}

// TestNewPullRequestFromRepository проверяет восстановление из БД
func TestNewPullRequestFromRepository(t *testing.T) {
	reviewers := []string{"reviewer-1", "reviewer-2"}
	mergedAt := func() *time.Time {
		t := time.Now().UTC()
		return &t
	}()

	pr := NewPullRequestFromRepository(
		"pr-100",
		"Legacy PR",
		"author-1",
		PRStatusMerged,
		reviewers,
		time.Now().Add(-24*time.Hour).UTC(),
		mergedAt,
	)

	if pr.ID() != "pr-100" {
		t.Errorf("ID = %v, want pr-100", pr.ID())
	}
	if pr.Status() != PRStatusMerged {
		t.Errorf("Status = %v, want MERGED", pr.Status())
	}
	if len(pr.AssignedReviewers()) != MaxReviewersCount {
		t.Errorf("AssignedReviewers length = %d, want %d", len(pr.AssignedReviewers()), MaxReviewersCount)
	}
	if pr.MergedAt() == nil {
		t.Errorf("MergedAt = nil, want non-nil")
	}
}
