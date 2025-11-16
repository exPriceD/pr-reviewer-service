package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/exPriceD/pr-reviewer-service/internal/domain/entity"
	"github.com/exPriceD/pr-reviewer-service/internal/domain/repository"
	repositorymocks "github.com/exPriceD/pr-reviewer-service/internal/domain/repository/mocks"
)

func TestReviewerSelector_SelectReviewers(t *testing.T) {
	tests := []struct {
		name          string
		teamName      string
		authorID      string
		setupMocks    func(*repositorymocks.MockUserRepository, *repositorymocks.MockPullRequestRepository)
		expectErr     bool
		expectedCount int
	}{
		{
			name:     "success - select 2 reviewers",
			teamName: "team-1",
			authorID: "author-1",
			setupMocks: func(userRepo *repositorymocks.MockUserRepository, prRepo *repositorymocks.MockPullRequestRepository) {
				userRepo.EXPECT().FindActiveByTeamName(gomock.Any(), "team-1").Return([]*entity.User{
					entity.NewUserFromRepository("author-1", "Author", "team-1", true, time.Now(), time.Now()),
					entity.NewUserFromRepository("reviewer-1", "Reviewer 1", "team-1", true, time.Now(), time.Now()),
					entity.NewUserFromRepository("reviewer-2", "Reviewer 2", "team-1", true, time.Now(), time.Now()),
				}, nil)
				prRepo.EXPECT().CountActiveReviewsByUserIDs(gomock.Any(), gomock.Any()).Return(map[string]int{
					"reviewer-1": 1,
					"reviewer-2": 0,
				}, nil)
			},
			expectErr:     false,
			expectedCount: entity.MaxReviewersCount,
		},
		{
			name:     "success - only author in team",
			teamName: "team-1",
			authorID: "author-1",
			setupMocks: func(userRepo *repositorymocks.MockUserRepository, prRepo *repositorymocks.MockPullRequestRepository) {
				userRepo.EXPECT().FindActiveByTeamName(gomock.Any(), "team-1").Return([]*entity.User{
					entity.NewUserFromRepository("author-1", "Author", "team-1", true, time.Now(), time.Now()),
				}, nil)
			},
			expectErr:     false,
			expectedCount: 0,
		},
		{
			name:     "success - one reviewer available",
			teamName: "team-1",
			authorID: "author-1",
			setupMocks: func(userRepo *repositorymocks.MockUserRepository, prRepo *repositorymocks.MockPullRequestRepository) {
				userRepo.EXPECT().FindActiveByTeamName(gomock.Any(), "team-1").Return([]*entity.User{
					entity.NewUserFromRepository("author-1", "Author", "team-1", true, time.Now(), time.Now()),
					entity.NewUserFromRepository("reviewer-1", "Reviewer 1", "team-1", true, time.Now(), time.Now()),
				}, nil)
				prRepo.EXPECT().CountActiveReviewsByUserIDs(gomock.Any(), gomock.Any()).Return(map[string]int{
					"reviewer-1": 0,
				}, nil)
			},
			expectErr:     false,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo := repositorymocks.NewMockUserRepository(ctrl)
			prRepo := repositorymocks.NewMockPullRequestRepository(ctrl)

			selector := NewReviewerSelector(userRepo, prRepo)

			tt.setupMocks(userRepo, prRepo)

			result, err := selector.SelectReviewers(context.Background(), tt.teamName, tt.authorID)

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if result != nil {
					t.Errorf("expected nil result, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Fatal("expected result, got nil")
				}
				if len(result) != tt.expectedCount {
					t.Errorf("expected %d reviewers, got %d", tt.expectedCount, len(result))
				}
				for _, reviewerID := range result {
					if reviewerID == tt.authorID {
						t.Errorf("author %s should not be in reviewers list", tt.authorID)
					}
				}
			}
		})
	}
}

func TestReviewerSelector_SelectReplacement(t *testing.T) {
	tests := []struct {
		name              string
		oldReviewerID     string
		authorID          string
		assignedReviewers []string
		setupMocks        func(*repositorymocks.MockUserRepository, *repositorymocks.MockPullRequestRepository)
		expectErr         bool
		expectedErr       error
	}{
		{
			name:              "success - select replacement",
			oldReviewerID:     "reviewer-1",
			authorID:          "author-1",
			assignedReviewers: []string{"reviewer-1", "reviewer-2"},
			setupMocks: func(userRepo *repositorymocks.MockUserRepository, prRepo *repositorymocks.MockPullRequestRepository) {
				userRepo.EXPECT().FindByID(gomock.Any(), "reviewer-1").Return(
					entity.NewUserFromRepository("reviewer-1", "Reviewer 1", "team-1", true, time.Now(), time.Now()),
					nil,
				)
				userRepo.EXPECT().FindActiveByTeamName(gomock.Any(), "team-1").Return([]*entity.User{
					entity.NewUserFromRepository("reviewer-1", "Reviewer 1", "team-1", true, time.Now(), time.Now()),
					entity.NewUserFromRepository("reviewer-2", "Reviewer 2", "team-1", true, time.Now(), time.Now()),
					entity.NewUserFromRepository("reviewer-3", "Reviewer 3", "team-1", true, time.Now(), time.Now()),
				}, nil)
				prRepo.EXPECT().CountActiveReviewsByUserIDs(gomock.Any(), gomock.Any()).Return(map[string]int{
					"reviewer-3": 0,
				}, nil)
			},
			expectErr: false,
		},
		{
			name:              "error - no active candidates",
			oldReviewerID:     "reviewer-1",
			authorID:          "author-1",
			assignedReviewers: []string{"reviewer-1"},
			setupMocks: func(userRepo *repositorymocks.MockUserRepository, prRepo *repositorymocks.MockPullRequestRepository) {
				userRepo.EXPECT().FindByID(gomock.Any(), "reviewer-1").Return(
					entity.NewUserFromRepository("reviewer-1", "Reviewer 1", "team-1", true, time.Now(), time.Now()),
					nil,
				)
				userRepo.EXPECT().FindActiveByTeamName(gomock.Any(), "team-1").Return([]*entity.User{
					entity.NewUserFromRepository("reviewer-1", "Reviewer 1", "team-1", true, time.Now(), time.Now()),
					entity.NewUserFromRepository("author-1", "Author", "team-1", true, time.Now(), time.Now()),
				}, nil)
			},
			expectErr:   true,
			expectedErr: ErrNoActiveCandidates,
		},
		{
			name:              "error - old reviewer not found",
			oldReviewerID:     "reviewer-1",
			authorID:          "author-1",
			assignedReviewers: []string{"reviewer-1"},
			setupMocks: func(userRepo *repositorymocks.MockUserRepository, prRepo *repositorymocks.MockPullRequestRepository) {
				userRepo.EXPECT().FindByID(gomock.Any(), "reviewer-1").Return(nil, repository.ErrNotFound)
			},
			expectErr: true,
		},
		{
			name:              "error - old reviewer FindByID error",
			oldReviewerID:     "reviewer-1",
			authorID:          "author-1",
			assignedReviewers: []string{"reviewer-1"},
			setupMocks: func(userRepo *repositorymocks.MockUserRepository, prRepo *repositorymocks.MockPullRequestRepository) {
				userRepo.EXPECT().FindByID(gomock.Any(), "reviewer-1").Return(nil, errors.New("database error"))
			},
			expectErr: true,
		},
		{
			name:              "error - FindActiveByTeamName error",
			oldReviewerID:     "reviewer-1",
			authorID:          "author-1",
			assignedReviewers: []string{"reviewer-1"},
			setupMocks: func(userRepo *repositorymocks.MockUserRepository, prRepo *repositorymocks.MockPullRequestRepository) {
				userRepo.EXPECT().FindByID(gomock.Any(), "reviewer-1").Return(
					entity.NewUserFromRepository("reviewer-1", "Reviewer 1", "team-1", true, time.Now(), time.Now()),
					nil,
				)
				userRepo.EXPECT().FindActiveByTeamName(gomock.Any(), "team-1").Return(nil, errors.New("database error"))
			},
			expectErr: true,
		},
		{
			name:              "error - CountActiveReviewsByUserIDs error",
			oldReviewerID:     "reviewer-1",
			authorID:          "author-1",
			assignedReviewers: []string{"reviewer-1"},
			setupMocks: func(userRepo *repositorymocks.MockUserRepository, prRepo *repositorymocks.MockPullRequestRepository) {
				userRepo.EXPECT().FindByID(gomock.Any(), "reviewer-1").Return(
					entity.NewUserFromRepository("reviewer-1", "Reviewer 1", "team-1", true, time.Now(), time.Now()),
					nil,
				)
				userRepo.EXPECT().FindActiveByTeamName(gomock.Any(), "team-1").Return([]*entity.User{
					entity.NewUserFromRepository("reviewer-1", "Reviewer 1", "team-1", true, time.Now(), time.Now()),
					entity.NewUserFromRepository("reviewer-2", "Reviewer 2", "team-1", true, time.Now(), time.Now()),
					entity.NewUserFromRepository("reviewer-3", "Reviewer 3", "team-1", true, time.Now(), time.Now()),
				}, nil)
				prRepo.EXPECT().CountActiveReviewsByUserIDs(gomock.Any(), gomock.Any()).Return(nil, errors.New("database error"))
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo := repositorymocks.NewMockUserRepository(ctrl)
			prRepo := repositorymocks.NewMockPullRequestRepository(ctrl)

			selector := NewReviewerSelector(userRepo, prRepo)

			tt.setupMocks(userRepo, prRepo)

			result, err := selector.SelectReplacement(context.Background(), tt.oldReviewerID, tt.authorID, tt.assignedReviewers)

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
				if result != "" {
					t.Errorf("expected empty result, got %s", result)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == "" {
					t.Error("expected replacement reviewer ID, got empty")
				}
				if result == tt.oldReviewerID {
					t.Errorf("replacement should not be the same as old reviewer")
				}
				if result == tt.authorID {
					t.Errorf("replacement should not be the author")
				}
			}
		})
	}
}
