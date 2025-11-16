package repository

import (
	"context"

	"github.com/exPriceD/pr-reviewer-service/internal/domain/entity"
)

type TeamRepository interface {
	Create(ctx context.Context, team *entity.Team) error
	FindByName(ctx context.Context, name string) (*entity.Team, error)
	Update(ctx context.Context, team *entity.Team) error
	Delete(ctx context.Context, name string) error
	Exists(ctx context.Context, name string) (bool, error)
}
