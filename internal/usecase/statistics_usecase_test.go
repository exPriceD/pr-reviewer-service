package usecase

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/exPriceD/pr-reviewer-service/internal/domain/entity"
	loggermocks "github.com/exPriceD/pr-reviewer-service/internal/domain/logger/mocks"
	repositorymocks "github.com/exPriceD/pr-reviewer-service/internal/domain/repository/mocks"
)

func TestStatisticsUseCase_GetStatistics(t *testing.T) {
	tests := []struct {
		name       string
		teamName   string
		setupMocks func(*repositorymocks.MockPullRequestRepository, *repositorymocks.MockUserRepository, *loggermocks.MockLogger)
		expectErr  bool
	}{
		{
			name:     "success - get statistics for team",
			teamName: "team-1",
			setupMocks: func(prRepo *repositorymocks.MockPullRequestRepository, userRepo *repositorymocks.MockUserRepository, logger *loggermocks.MockLogger) {
				prRepo.EXPECT().GetStats(gomock.Any()).Return(10, 5, 5, nil)
				userRepo.EXPECT().FindByTeamName(gomock.Any(), "team-1").Return([]*entity.User{
					entity.NewUserFromRepository("user-1", "User 1", "team-1", true, time.Now(), time.Now()),
					entity.NewUserFromRepository("user-2", "User 2", "team-1", true, time.Now(), time.Now()),
				}, nil)
				prRepo.EXPECT().CountReviewsByUserIDs(gomock.Any(), gomock.Any()).Return(map[string]int{
					"user-1": 3,
					"user-2": 2,
				}, nil)
				prRepo.EXPECT().CountActiveReviewsByUserIDs(gomock.Any(), gomock.Any()).Return(map[string]int{
					"user-1": 1,
					"user-2": 1,
				}, nil)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr: false,
		},
		{
			name:     "success - get statistics for all teams",
			teamName: "",
			setupMocks: func(prRepo *repositorymocks.MockPullRequestRepository, userRepo *repositorymocks.MockUserRepository, logger *loggermocks.MockLogger) {
				prRepo.EXPECT().GetStats(gomock.Any()).Return(10, 5, 5, nil)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr: false,
		},
		{
			name:     "success - team with no users",
			teamName: "team-1",
			setupMocks: func(prRepo *repositorymocks.MockPullRequestRepository, userRepo *repositorymocks.MockUserRepository, logger *loggermocks.MockLogger) {
				prRepo.EXPECT().GetStats(gomock.Any()).Return(10, 5, 5, nil)
				userRepo.EXPECT().FindByTeamName(gomock.Any(), "team-1").Return([]*entity.User{}, nil)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			prRepo := repositorymocks.NewMockPullRequestRepository(ctrl)
			userRepo := repositorymocks.NewMockUserRepository(ctrl)
			logger := loggermocks.NewMockLogger(ctrl)

			uc := NewStatisticsUseCase(prRepo, userRepo, logger)

			tt.setupMocks(prRepo, userRepo, logger)

			result, err := uc.GetStatistics(context.Background(), tt.teamName)

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
				if result.PRStats.Total != 10 {
					t.Errorf("expected total PRs 10, got %d", result.PRStats.Total)
				}
			}
		})
	}
}
