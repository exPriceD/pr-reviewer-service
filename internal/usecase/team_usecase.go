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

// TeamUseCase Use Case для работы с командами
type TeamUseCase struct {
	txManager transaction.Manager
	teamRepo  repository.TeamRepository
	userRepo  repository.UserRepository
	logger    logger.Logger
}

// NewTeamUseCase создает новый TeamUseCase
func NewTeamUseCase(
	txManager transaction.Manager,
	teamRepo repository.TeamRepository,
	userRepo repository.UserRepository,
	logger logger.Logger,
) *TeamUseCase {
	return &TeamUseCase{
		txManager: txManager,
		teamRepo:  teamRepo,
		userRepo:  userRepo,
		logger:    logger,
	}
}

// CreateTeam создает команду с участниками (создает/обновляет пользователей)
// POST /team/add
func (uc *TeamUseCase) CreateTeam(ctx context.Context, req dto.CreateTeamRequest) (*dto.TeamDTO, error) {
	uc.logger.Info("Creating team", "team_name", req.TeamName, "members_count", len(req.Members))

	exists, err := uc.teamRepo.Exists(ctx, req.TeamName)
	if err != nil {
		uc.logger.Error("Failed to check team existence", "error", err)
		return nil, fmt.Errorf("failed to check team existence: %w", err)
	}
	if exists {
		return nil, ErrTeamAlreadyExists
	}

	var team *entity.Team
	var users []*entity.User

	err = uc.txManager.Do(ctx, func(ctx context.Context) error {
		team, err = entity.NewTeam(req.TeamName)
		if err != nil {
			return fmt.Errorf("failed to create team entity: %w", err)
		}

		if err := uc.teamRepo.Create(ctx, team); err != nil {
			return fmt.Errorf("failed to save team: %w", err)
		}

		for _, memberReq := range req.Members {
			existingUser, err := uc.userRepo.FindByID(ctx, memberReq.UserID)
			if err != nil && !errors.Is(err, repository.ErrNotFound) {
				return fmt.Errorf("failed to check user existence %s: %w", memberReq.UserID, err)
			}

			if existingUser != nil {
				if err := existingUser.ChangeTeam(req.TeamName); err != nil {
					return fmt.Errorf("failed to change team for user %s: %w", memberReq.UserID, err)
				}
				if memberReq.IsActive && !existingUser.IsActive() {
					existingUser.Activate()
				} else if !memberReq.IsActive && existingUser.IsActive() {
					existingUser.Deactivate()
				}
				if err := uc.userRepo.Update(ctx, existingUser); err != nil {
					return fmt.Errorf("failed to update user %s: %w", memberReq.UserID, err)
				}
				users = append(users, existingUser)
			} else {
				user, err := entity.NewUser(memberReq.UserID, memberReq.Username, req.TeamName)
				if err != nil {
					return fmt.Errorf("failed to create user entity %s: %w", memberReq.UserID, err)
				}
				if !memberReq.IsActive {
					user.Deactivate()
				}
				if err := uc.userRepo.Create(ctx, user); err != nil {
					return fmt.Errorf("failed to save user %s: %w", memberReq.UserID, err)
				}
				users = append(users, user)
			}
		}
		return nil
	})
	if err != nil {
		uc.logger.Error("Failed to create team", "error", err, "team_name", req.TeamName)
		return nil, err
	}

	uc.logger.Info("Team created successfully", "team_name", req.TeamName, "members_count", len(users))
	result := dto.ToTeamDTO(team, users)
	return &result, nil
}

// GetTeam получает команду с участниками
// GET /team/get?team_name=
func (uc *TeamUseCase) GetTeam(ctx context.Context, teamName string) (*dto.TeamDTO, error) {
	uc.logger.Info("Getting team", "team_name", teamName)

	team, err := uc.teamRepo.FindByName(ctx, teamName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrTeamNotFound
		}
		uc.logger.Error("Failed to find team", "error", err, "team_name", teamName)
		return nil, fmt.Errorf("failed to find team: %w", err)
	}

	users, err := uc.userRepo.FindByTeamName(ctx, teamName)
	if err != nil {
		uc.logger.Error("Failed to find team users", "error", err, "team_name", teamName)
		return nil, fmt.Errorf("failed to find team users: %w", err)
	}

	uc.logger.Info("Team retrieved successfully", "team_name", teamName, "members_count", len(users))
	result := dto.ToTeamDTO(team, users)
	return &result, nil
}

// DeactivateTeamMembers массово деактивирует всех пользователей команды
// Примечание: согласно ТЗ, должна быть безопасная переназначаемость открытых PR
// (стремиться уложиться в 100 мс для средних объёмов данных)
func (uc *TeamUseCase) DeactivateTeamMembers(ctx context.Context, teamName string) (*dto.TeamDTO, error) {
	uc.logger.Info("Deactivating team members", "team_name", teamName)

	team, err := uc.teamRepo.FindByName(ctx, teamName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrTeamNotFound
		}
		uc.logger.Error("Failed to find team", "error", err, "team_name", teamName)
		return nil, fmt.Errorf("failed to find team: %w", err)
	}

	users, err := uc.userRepo.FindByTeamName(ctx, teamName)
	if err != nil {
		uc.logger.Error("Failed to find team users", "error", err, "team_name", teamName)
		return nil, fmt.Errorf("failed to find team users: %w", err)
	}

	if len(users) == 0 {
		uc.logger.Info("No users to deactivate", "team_name", teamName)
		result := dto.ToTeamDTO(team, users)
		return &result, nil
	}

	err = uc.txManager.Do(ctx, func(ctx context.Context) error {
		if err := uc.userRepo.BatchDeactivateByTeamName(ctx, teamName); err != nil {
			return fmt.Errorf("failed to batch deactivate team members: %w", err)
		}
		return nil
	})
	if err != nil {
		uc.logger.Error("Failed to deactivate team members", "error", err, "team_name", teamName)
		return nil, err
	}

	users, err = uc.userRepo.FindByTeamName(ctx, teamName)
	if err != nil {
		uc.logger.Error("Failed to reload team users", "error", err, "team_name", teamName)
		return nil, fmt.Errorf("failed to reload team users: %w", err)
	}

	uc.logger.Info("Team members deactivated successfully", "team_name", teamName, "members_count", len(users))
	result := dto.ToTeamDTO(team, users)
	return &result, nil
}
