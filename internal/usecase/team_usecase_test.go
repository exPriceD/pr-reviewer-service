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

func TestTeamUseCase_CreateTeam(t *testing.T) {
	tests := []struct {
		name        string
		req         dto.CreateTeamRequest
		setupMocks  func(*repositorymocks.MockTeamRepository, *repositorymocks.MockUserRepository, *transactionmocks.MockManager, *loggermocks.MockLogger)
		expectErr   bool
		expectedErr error
	}{
		{
			name: "success - create team with new users",
			req: dto.CreateTeamRequest{
				TeamName: "team-1",
				Members: []dto.TeamMemberRequest{
					{UserID: "user-1", Username: "User 1", IsActive: true},
					{UserID: "user-2", Username: "User 2", IsActive: true},
				},
			},
			setupMocks: func(teamRepo *repositorymocks.MockTeamRepository, userRepo *repositorymocks.MockUserRepository, txManager *transactionmocks.MockManager, logger *loggermocks.MockLogger) {
				teamRepo.EXPECT().Exists(gomock.Any(), "team-1").Return(false, nil)
				txManager.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				teamRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
				userRepo.EXPECT().FindByID(gomock.Any(), "user-1").Return(nil, repository.ErrNotFound)
				userRepo.EXPECT().FindByID(gomock.Any(), "user-2").Return(nil, repository.ErrNotFound)
				userRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil).Times(2)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr: false,
		},
		{
			name: "success - create team with existing users",
			req: dto.CreateTeamRequest{
				TeamName: "team-1",
				Members: []dto.TeamMemberRequest{
					{UserID: "user-1", Username: "User 1", IsActive: true},
				},
			},
			setupMocks: func(teamRepo *repositorymocks.MockTeamRepository, userRepo *repositorymocks.MockUserRepository, txManager *transactionmocks.MockManager, logger *loggermocks.MockLogger) {
				teamRepo.EXPECT().Exists(gomock.Any(), "team-1").Return(false, nil)
				txManager.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				teamRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
				userRepo.EXPECT().FindByID(gomock.Any(), "user-1").Return(
					entity.NewUserFromRepository("user-1", "User 1", "old-team", true, time.Now(), time.Now()),
					nil,
				)
				userRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr: false,
		},
		{
			name: "error - team already exists",
			req: dto.CreateTeamRequest{
				TeamName: "team-1",
				Members:  []dto.TeamMemberRequest{},
			},
			setupMocks: func(teamRepo *repositorymocks.MockTeamRepository, userRepo *repositorymocks.MockUserRepository, txManager *transactionmocks.MockManager, logger *loggermocks.MockLogger) {
				teamRepo.EXPECT().Exists(gomock.Any(), "team-1").Return(true, nil)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr:   true,
			expectedErr: ErrTeamAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			teamRepo := repositorymocks.NewMockTeamRepository(ctrl)
			userRepo := repositorymocks.NewMockUserRepository(ctrl)
			txManager := transactionmocks.NewMockManager(ctrl)
			logger := loggermocks.NewMockLogger(ctrl)

			uc := NewTeamUseCase(txManager, teamRepo, userRepo, logger)

			tt.setupMocks(teamRepo, userRepo, txManager, logger)

			result, err := uc.CreateTeam(context.Background(), tt.req)

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
				if result.TeamName != tt.req.TeamName {
					t.Errorf("expected team name %s, got %s", tt.req.TeamName, result.TeamName)
				}
			}
		})
	}
}

func TestTeamUseCase_GetTeam(t *testing.T) {
	tests := []struct {
		name        string
		teamName    string
		setupMocks  func(*repositorymocks.MockTeamRepository, *repositorymocks.MockUserRepository, *loggermocks.MockLogger)
		expectErr   bool
		expectedErr error
	}{
		{
			name:     "success - get team",
			teamName: "team-1",
			setupMocks: func(teamRepo *repositorymocks.MockTeamRepository, userRepo *repositorymocks.MockUserRepository, logger *loggermocks.MockLogger) {
				now := time.Now()
				teamRepo.EXPECT().FindByName(gomock.Any(), "team-1").Return(
					entity.NewTeamFromRepository("team-1", now, now),
					nil,
				)
				userRepo.EXPECT().FindByTeamName(gomock.Any(), "team-1").Return([]*entity.User{
					entity.NewUserFromRepository("user-1", "User 1", "team-1", true, time.Now(), time.Now()),
				}, nil)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr: false,
		},
		{
			name:     "error - team not found",
			teamName: "team-1",
			setupMocks: func(teamRepo *repositorymocks.MockTeamRepository, userRepo *repositorymocks.MockUserRepository, logger *loggermocks.MockLogger) {
				teamRepo.EXPECT().FindByName(gomock.Any(), "team-1").Return(nil, repository.ErrNotFound)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
				logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr:   true,
			expectedErr: ErrTeamNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			teamRepo := repositorymocks.NewMockTeamRepository(ctrl)
			userRepo := repositorymocks.NewMockUserRepository(ctrl)
			txManager := transactionmocks.NewMockManager(ctrl)
			logger := loggermocks.NewMockLogger(ctrl)

			uc := NewTeamUseCase(txManager, teamRepo, userRepo, logger)

			tt.setupMocks(teamRepo, userRepo, logger)

			result, err := uc.GetTeam(context.Background(), tt.teamName)

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
				if result.TeamName != tt.teamName {
					t.Errorf("expected team name %s, got %s", tt.teamName, result.TeamName)
				}
			}
		})
	}
}

func TestTeamUseCase_DeactivateTeamMembers(t *testing.T) {
	tests := []struct {
		name        string
		teamName    string
		setupMocks  func(*repositorymocks.MockTeamRepository, *repositorymocks.MockUserRepository, *transactionmocks.MockManager, *loggermocks.MockLogger)
		expectErr   bool
		expectedErr error
	}{
		{
			name:     "success - deactivate team members",
			teamName: "team-1",
			setupMocks: func(teamRepo *repositorymocks.MockTeamRepository, userRepo *repositorymocks.MockUserRepository, txManager *transactionmocks.MockManager, logger *loggermocks.MockLogger) {
				now := time.Now()
				teamRepo.EXPECT().FindByName(gomock.Any(), "team-1").Return(
					entity.NewTeamFromRepository("team-1", now, now),
					nil,
				)
				userRepo.EXPECT().FindByTeamName(gomock.Any(), "team-1").Return([]*entity.User{
					entity.NewUserFromRepository("user-1", "User 1", "team-1", true, time.Now(), time.Now()),
				}, nil).Times(2)
				txManager.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				userRepo.EXPECT().BatchDeactivateByTeamName(gomock.Any(), "team-1").Return(nil)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr: false,
		},
		{
			name:     "error - team not found",
			teamName: "team-1",
			setupMocks: func(teamRepo *repositorymocks.MockTeamRepository, userRepo *repositorymocks.MockUserRepository, txManager *transactionmocks.MockManager, logger *loggermocks.MockLogger) {
				teamRepo.EXPECT().FindByName(gomock.Any(), "team-1").Return(nil, repository.ErrNotFound)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
				logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectErr:   true,
			expectedErr: ErrTeamNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			teamRepo := repositorymocks.NewMockTeamRepository(ctrl)
			userRepo := repositorymocks.NewMockUserRepository(ctrl)
			txManager := transactionmocks.NewMockManager(ctrl)
			logger := loggermocks.NewMockLogger(ctrl)

			uc := NewTeamUseCase(txManager, teamRepo, userRepo, logger)

			tt.setupMocks(teamRepo, userRepo, txManager, logger)

			result, err := uc.DeactivateTeamMembers(context.Background(), tt.teamName)

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
				if result.TeamName != tt.teamName {
					t.Errorf("expected team name %s, got %s", tt.teamName, result.TeamName)
				}
			}
		})
	}
}
