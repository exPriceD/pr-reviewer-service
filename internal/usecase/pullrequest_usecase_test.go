package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/exPriceD/pr-reviewer-service/internal/domain/entity"
	loggermocks "github.com/exPriceD/pr-reviewer-service/internal/domain/logger/mocks"
	"github.com/exPriceD/pr-reviewer-service/internal/domain/repository"
	repositorymocks "github.com/exPriceD/pr-reviewer-service/internal/domain/repository/mocks"
	transactionmocks "github.com/exPriceD/pr-reviewer-service/internal/domain/transaction/mocks"
	"github.com/exPriceD/pr-reviewer-service/internal/usecase/dto"
)

func TestPullRequestUseCase_CreatePR(t *testing.T) {
	tests := []struct {
		name          string
		req           dto.CreatePRRequest
		setupMocks    func(*repositorymocks.MockPullRequestRepository, *repositorymocks.MockUserRepository, *transactionmocks.MockManager, *loggermocks.MockLogger)
		expectErr     bool
		expectedErr   error
		expectedCount int
	}{
		{
			name: "success - PR created with 2 reviewers",
			req: dto.CreatePRRequest{
				PullRequestID:   "pr-1",
				PullRequestName: "Test PR",
				AuthorID:        "author-1",
			},
			setupMocks: func(prRepo *repositorymocks.MockPullRequestRepository, userRepo *repositorymocks.MockUserRepository, txManager *transactionmocks.MockManager, logger *loggermocks.MockLogger) {
				prRepo.EXPECT().Exists(gomock.Any(), "pr-1").Return(false, nil)
				userRepo.EXPECT().FindByID(gomock.Any(), "author-1").Return(
					entity.NewUserFromRepository("author-1", "Author", "team-1", true, time.Now(), time.Now()),
					nil,
				)
				txManager.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				userRepo.EXPECT().FindActiveByTeamName(gomock.Any(), "team-1").Return([]*entity.User{
					entity.NewUserFromRepository("reviewer-1", "Reviewer 1", "team-1", true, time.Now(), time.Now()),
					entity.NewUserFromRepository("reviewer-2", "Reviewer 2", "team-1", true, time.Now(), time.Now()),
				}, nil)
				prRepo.EXPECT().CountActiveReviewsByUserIDs(gomock.Any(), gomock.Any()).Return(map[string]int{
					"reviewer-1": 1,
					"reviewer-2": 0,
				}, nil)
				prRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr:     false,
			expectedCount: entity.MaxReviewersCount,
		},
		{
			name: "error - PR already exists",
			req: dto.CreatePRRequest{
				PullRequestID:   "pr-1",
				PullRequestName: "Test PR",
				AuthorID:        "author-1",
			},
			setupMocks: func(prRepo *repositorymocks.MockPullRequestRepository, userRepo *repositorymocks.MockUserRepository, txManager *transactionmocks.MockManager, logger *loggermocks.MockLogger) {
				prRepo.EXPECT().Exists(gomock.Any(), "pr-1").Return(true, nil)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr:   true,
			expectedErr: ErrPRAlreadyExists,
		},
		{
			name: "error - author not found",
			req: dto.CreatePRRequest{
				PullRequestID:   "pr-1",
				PullRequestName: "Test PR",
				AuthorID:        "author-1",
			},
			setupMocks: func(prRepo *repositorymocks.MockPullRequestRepository, userRepo *repositorymocks.MockUserRepository, txManager *transactionmocks.MockManager, logger *loggermocks.MockLogger) {
				prRepo.EXPECT().Exists(gomock.Any(), "pr-1").Return(false, nil)
				userRepo.EXPECT().FindByID(gomock.Any(), "author-1").Return(nil, repository.ErrNotFound)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
				logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr:   true,
			expectedErr: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			prRepo := repositorymocks.NewMockPullRequestRepository(ctrl)
			userRepo := repositorymocks.NewMockUserRepository(ctrl)
			txManager := transactionmocks.NewMockManager(ctrl)
			logger := loggermocks.NewMockLogger(ctrl)
			reviewerSelector := NewReviewerSelector(userRepo, prRepo)

			uc := NewPullRequestUseCase(txManager, prRepo, userRepo, reviewerSelector, logger)

			tt.setupMocks(prRepo, userRepo, txManager, logger)

			result, err := uc.CreatePR(context.Background(), tt.req)

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
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
				if result.PullRequestID != tt.req.PullRequestID {
					t.Errorf("expected PR ID %s, got %s", tt.req.PullRequestID, result.PullRequestID)
				}
				if len(result.AssignedReviewers) != tt.expectedCount {
					t.Errorf("expected %d reviewers, got %d", tt.expectedCount, len(result.AssignedReviewers))
				}
			}
		})
	}
}

func TestPullRequestUseCase_MergePR(t *testing.T) {
	tests := []struct {
		name        string
		prID        string
		setupMocks  func(*repositorymocks.MockPullRequestRepository, *loggermocks.MockLogger)
		expectErr   bool
		expectedErr error
	}{
		{
			name: "success - PR merged",
			prID: "pr-1",
			setupMocks: func(prRepo *repositorymocks.MockPullRequestRepository, logger *loggermocks.MockLogger) {
				prRepo.EXPECT().MergePR(gomock.Any(), "pr-1").Return(nil)
				prRepo.EXPECT().FindByID(gomock.Any(), "pr-1").Return(
					entity.NewPullRequestFromRepository("pr-1", "Test PR", "author-1", entity.PRStatusMerged, []string{}, time.Now(), timePtr(time.Now())),
					nil,
				)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr: false,
		},
		{
			name: "success - PR already merged (idempotent)",
			prID: "pr-1",
			setupMocks: func(prRepo *repositorymocks.MockPullRequestRepository, logger *loggermocks.MockLogger) {
				prRepo.EXPECT().MergePR(gomock.Any(), "pr-1").Return(repository.ErrNotFound)
				prRepo.EXPECT().FindByID(gomock.Any(), "pr-1").Return(
					entity.NewPullRequestFromRepository("pr-1", "Test PR", "author-1", entity.PRStatusMerged, []string{}, time.Now(), timePtr(time.Now())),
					nil,
				)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr: false,
		},
		{
			name: "error - PR not found",
			prID: "pr-1",
			setupMocks: func(prRepo *repositorymocks.MockPullRequestRepository, logger *loggermocks.MockLogger) {
				prRepo.EXPECT().MergePR(gomock.Any(), "pr-1").Return(repository.ErrNotFound)
				prRepo.EXPECT().FindByID(gomock.Any(), "pr-1").Return(nil, repository.ErrNotFound)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
				logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr:   true,
			expectedErr: ErrPRNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			prRepo := repositorymocks.NewMockPullRequestRepository(ctrl)
			userRepo := repositorymocks.NewMockUserRepository(ctrl)
			txManager := transactionmocks.NewMockManager(ctrl)
			logger := loggermocks.NewMockLogger(ctrl)
			reviewerSelector := NewReviewerSelector(userRepo, prRepo)

			uc := NewPullRequestUseCase(txManager, prRepo, userRepo, reviewerSelector, logger)

			tt.setupMocks(prRepo, logger)

			result, err := uc.MergePR(context.Background(), tt.prID)

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
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
				if result.Status != string(entity.PRStatusMerged) {
					t.Errorf("expected status MERGED, got %s", result.Status)
				}
			}
		})
	}
}

func TestPullRequestUseCase_ReassignReviewer(t *testing.T) {
	tests := []struct {
		name        string
		req         dto.ReassignReviewerRequest
		setupMocks  func(*repositorymocks.MockPullRequestRepository, *repositorymocks.MockUserRepository, *transactionmocks.MockManager, *loggermocks.MockLogger)
		expectErr   bool
		expectedErr error
	}{
		{
			name: "success - reviewer reassigned",
			req: dto.ReassignReviewerRequest{
				PullRequestID: "pr-1",
				OldUserID:     "reviewer-1",
			},
			setupMocks: func(prRepo *repositorymocks.MockPullRequestRepository, userRepo *repositorymocks.MockUserRepository, txManager *transactionmocks.MockManager, logger *loggermocks.MockLogger) {
				txManager.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				prRepo.EXPECT().FindByIDForUpdate(gomock.Any(), "pr-1").Return(
					entity.NewPullRequestFromRepository("pr-1", "Test PR", "author-1", entity.PRStatusOpen, []string{"reviewer-1"}, time.Now(), nil),
					nil,
				)
				userRepo.EXPECT().FindByID(gomock.Any(), "reviewer-1").Return(
					entity.NewUserFromRepository("reviewer-1", "Reviewer 1", "team-1", true, time.Now(), time.Now()),
					nil,
				)
				userRepo.EXPECT().FindActiveByTeamName(gomock.Any(), "team-1").Return([]*entity.User{
					entity.NewUserFromRepository("reviewer-2", "Reviewer 2", "team-1", true, time.Now(), time.Now()),
				}, nil)
				prRepo.EXPECT().CountActiveReviewsByUserIDs(gomock.Any(), gomock.Any()).Return(map[string]int{
					"reviewer-2": 0,
				}, nil)
				prRepo.EXPECT().ReplaceReviewer(gomock.Any(), "pr-1", "reviewer-1", "reviewer-2").Return(nil)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr: false,
		},
		{
			name: "error - PR not found",
			req: dto.ReassignReviewerRequest{
				PullRequestID: "pr-1",
				OldUserID:     "reviewer-1",
			},
			setupMocks: func(prRepo *repositorymocks.MockPullRequestRepository, userRepo *repositorymocks.MockUserRepository, txManager *transactionmocks.MockManager, logger *loggermocks.MockLogger) {
				txManager.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				prRepo.EXPECT().FindByIDForUpdate(gomock.Any(), "pr-1").Return(nil, repository.ErrNotFound)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
				logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr:   true,
			expectedErr: ErrPRNotFound,
		},
		{
			name: "error - PR already merged",
			req: dto.ReassignReviewerRequest{
				PullRequestID: "pr-1",
				OldUserID:     "reviewer-1",
			},
			setupMocks: func(prRepo *repositorymocks.MockPullRequestRepository, userRepo *repositorymocks.MockUserRepository, txManager *transactionmocks.MockManager, logger *loggermocks.MockLogger) {
				txManager.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				prRepo.EXPECT().FindByIDForUpdate(gomock.Any(), "pr-1").Return(
					entity.NewPullRequestFromRepository("pr-1", "Test PR", "author-1", entity.PRStatusMerged, []string{"reviewer-1"}, time.Now(), timePtr(time.Now())),
					nil,
				)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
				logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr:   true,
			expectedErr: ErrPRAlreadyMerged,
		},
		{
			name: "error - reviewer not assigned",
			req: dto.ReassignReviewerRequest{
				PullRequestID: "pr-1",
				OldUserID:     "reviewer-1",
			},
			setupMocks: func(prRepo *repositorymocks.MockPullRequestRepository, userRepo *repositorymocks.MockUserRepository, txManager *transactionmocks.MockManager, logger *loggermocks.MockLogger) {
				txManager.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				prRepo.EXPECT().FindByIDForUpdate(gomock.Any(), "pr-1").Return(
					entity.NewPullRequestFromRepository("pr-1", "Test PR", "author-1", entity.PRStatusOpen, []string{"reviewer-2"}, time.Now(), nil),
					nil,
				)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
				logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr:   true,
			expectedErr: ErrReviewerNotAssigned,
		},
		{
			name: "error - no active candidates",
			req: dto.ReassignReviewerRequest{
				PullRequestID: "pr-1",
				OldUserID:     "reviewer-1",
			},
			setupMocks: func(prRepo *repositorymocks.MockPullRequestRepository, userRepo *repositorymocks.MockUserRepository, txManager *transactionmocks.MockManager, logger *loggermocks.MockLogger) {
				txManager.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				prRepo.EXPECT().FindByIDForUpdate(gomock.Any(), "pr-1").Return(
					entity.NewPullRequestFromRepository("pr-1", "Test PR", "author-1", entity.PRStatusOpen, []string{"reviewer-1"}, time.Now(), nil),
					nil,
				)
				userRepo.EXPECT().FindByID(gomock.Any(), "reviewer-1").Return(
					entity.NewUserFromRepository("reviewer-1", "Reviewer 1", "team-1", true, time.Now(), time.Now()),
					nil,
				)
				userRepo.EXPECT().FindActiveByTeamName(gomock.Any(), "team-1").Return([]*entity.User{
					entity.NewUserFromRepository("reviewer-1", "Reviewer 1", "team-1", true, time.Now(), time.Now()),
					entity.NewUserFromRepository("author-1", "Author", "team-1", true, time.Now(), time.Now()),
				}, nil)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
				logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr:   true,
			expectedErr: ErrNoActiveCandidates,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			prRepo := repositorymocks.NewMockPullRequestRepository(ctrl)
			userRepo := repositorymocks.NewMockUserRepository(ctrl)
			txManager := transactionmocks.NewMockManager(ctrl)
			logger := loggermocks.NewMockLogger(ctrl)
			reviewerSelector := NewReviewerSelector(userRepo, prRepo)

			uc := NewPullRequestUseCase(txManager, prRepo, userRepo, reviewerSelector, logger)

			tt.setupMocks(prRepo, userRepo, txManager, logger)

			result, newReviewerID, err := uc.ReassignReviewer(context.Background(), tt.req)

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
				if result != nil {
					t.Errorf("expected nil result, got %v", result)
				}
				if newReviewerID != "" {
					t.Errorf("expected empty newReviewerID, got %s", newReviewerID)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Fatal("expected result, got nil")
				}
				if newReviewerID == "" {
					t.Error("expected newReviewerID, got empty")
				}
			}
		})
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func TestNewPullRequestUseCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prRepo := repositorymocks.NewMockPullRequestRepository(ctrl)
	userRepo := repositorymocks.NewMockUserRepository(ctrl)
	txManager := transactionmocks.NewMockManager(ctrl)
	logger := loggermocks.NewMockLogger(ctrl)
	reviewerSelector := NewReviewerSelector(userRepo, prRepo)

	uc := NewPullRequestUseCase(txManager, prRepo, userRepo, reviewerSelector, logger)

	if uc == nil {
		t.Fatal("expected non-nil use case")
	}
}
