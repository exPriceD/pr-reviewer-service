package repository

import (
	"context"

	"github.com/exPriceD/pr-reviewer-service/internal/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByTeamName(ctx context.Context, teamName string) ([]*entity.User, error)
	FindActiveByTeamName(ctx context.Context, teamName string) ([]*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	BatchDeactivateByTeamName(ctx context.Context, teamName string) error
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)
}
