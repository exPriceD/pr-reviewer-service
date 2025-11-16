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
	"github.com/exPriceD/pr-reviewer-service/internal/usecase/dto"
)

func TestUserUseCase_SetUserActive(t *testing.T) {
	tests := []struct {
		name        string
		req         dto.SetUserActiveRequest
		setupMocks  func(*repositorymocks.MockUserRepository, *loggermocks.MockLogger)
		expectErr   bool
		expectedErr error
	}{
		{
			name: "success - activate user",
			req: dto.SetUserActiveRequest{
				UserID:   "user-1",
				IsActive: true,
			},
			setupMocks: func(userRepo *repositorymocks.MockUserRepository, logger *loggermocks.MockLogger) {
				user := entity.NewUserFromRepository("user-1", "User 1", "team-1", false, time.Now(), time.Now())
				userRepo.EXPECT().FindByID(gomock.Any(), "user-1").Return(user, nil)
				userRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr: false,
		},
		{
			name: "success - deactivate user",
			req: dto.SetUserActiveRequest{
				UserID:   "user-1",
				IsActive: false,
			},
			setupMocks: func(userRepo *repositorymocks.MockUserRepository, logger *loggermocks.MockLogger) {
				user := entity.NewUserFromRepository("user-1", "User 1", "team-1", true, time.Now(), time.Now())
				userRepo.EXPECT().FindByID(gomock.Any(), "user-1").Return(user, nil)
				userRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr: false,
		},
		{
			name: "error - user not found",
			req: dto.SetUserActiveRequest{
				UserID:   "user-1",
				IsActive: true,
			},
			setupMocks: func(userRepo *repositorymocks.MockUserRepository, logger *loggermocks.MockLogger) {
				userRepo.EXPECT().FindByID(gomock.Any(), "user-1").Return(nil, repository.ErrNotFound)
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

			userRepo := repositorymocks.NewMockUserRepository(ctrl)
			prRepo := repositorymocks.NewMockPullRequestRepository(ctrl)
			logger := loggermocks.NewMockLogger(ctrl)

			uc := NewUserUseCase(userRepo, prRepo, logger)

			tt.setupMocks(userRepo, logger)

			result, err := uc.SetUserActive(context.Background(), tt.req)

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
				if result.UserID != tt.req.UserID {
					t.Errorf("expected user ID %s, got %s", tt.req.UserID, result.UserID)
				}
				if result.IsActive != tt.req.IsActive {
					t.Errorf("expected is_active %v, got %v", tt.req.IsActive, result.IsActive)
				}
			}
		})
	}
}

func TestUserUseCase_GetUserReviews(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		setupMocks  func(*repositorymocks.MockUserRepository, *repositorymocks.MockPullRequestRepository, *loggermocks.MockLogger)
		expectErr   bool
		expectedErr error
		expectedLen int
	}{
		{
			name:   "success - get user reviews",
			userID: "user-1",
			setupMocks: func(userRepo *repositorymocks.MockUserRepository, prRepo *repositorymocks.MockPullRequestRepository, logger *loggermocks.MockLogger) {
				userRepo.EXPECT().Exists(gomock.Any(), "user-1").Return(true, nil)
				prRepo.EXPECT().FindByReviewerID(gomock.Any(), "user-1").Return([]*entity.PullRequest{
					entity.NewPullRequestFromRepository("pr-1", "PR 1", "author-1", entity.PRStatusOpen, []string{"user-1"}, time.Now(), nil),
				}, nil)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr:   false,
			expectedLen: 1,
		},
		{
			name:   "error - user not found",
			userID: "user-1",
			setupMocks: func(userRepo *repositorymocks.MockUserRepository, prRepo *repositorymocks.MockPullRequestRepository, logger *loggermocks.MockLogger) {
				userRepo.EXPECT().Exists(gomock.Any(), "user-1").Return(false, nil)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr:   true,
			expectedErr: ErrUserNotFound,
		},
		{
			name:   "success - no reviews",
			userID: "user-1",
			setupMocks: func(userRepo *repositorymocks.MockUserRepository, prRepo *repositorymocks.MockPullRequestRepository, logger *loggermocks.MockLogger) {
				userRepo.EXPECT().Exists(gomock.Any(), "user-1").Return(true, nil)
				prRepo.EXPECT().FindByReviewerID(gomock.Any(), "user-1").Return([]*entity.PullRequest{}, nil)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr:   false,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo := repositorymocks.NewMockUserRepository(ctrl)
			prRepo := repositorymocks.NewMockPullRequestRepository(ctrl)
			logger := loggermocks.NewMockLogger(ctrl)

			uc := NewUserUseCase(userRepo, prRepo, logger)

			tt.setupMocks(userRepo, prRepo, logger)

			result, err := uc.GetUserReviews(context.Background(), tt.userID)

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
				if len(result) != tt.expectedLen {
					t.Errorf("expected %d reviews, got %d", tt.expectedLen, len(result))
				}
			}
		})
	}
}
