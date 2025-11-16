package team

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	trmsql "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"

	"github.com/exPriceD/pr-reviewer-service/internal/domain/entity"
	"github.com/exPriceD/pr-reviewer-service/internal/domain/repository"
)

var _ repository.TeamRepository = (*Repository)(nil)

type Repository struct {
	db     *sql.DB
	getter *trmsql.CtxGetter
}

func NewRepository(db *sql.DB, getter *trmsql.CtxGetter) *Repository {
	return &Repository{
		db:     db,
		getter: getter,
	}
}

// getDB возвращает *sql.DB или *sql.Tx в зависимости от контекста
func (r *Repository) getDB(ctx context.Context) interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
} {
	return r.getter.DefaultTrOrDB(ctx, r.db)
}

func (r *Repository) Create(ctx context.Context, team *entity.Team) error {
	model := FromEntity(team)

	query := `
		INSERT INTO teams (team_name, created_at, updated_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.getDB(ctx).ExecContext(
		ctx,
		query,
		model.Name,
		model.CreatedAt,
		model.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	return nil
}

func (r *Repository) FindByName(ctx context.Context, name string) (*entity.Team, error) {
	query := `
		SELECT team_name, created_at, updated_at
		FROM teams
		WHERE team_name = $1
	`

	var model Model
	err := r.getDB(ctx).QueryRowContext(ctx, query, name).Scan(
		&model.Name,
		&model.CreatedAt,
		&model.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find team: %w", err)
	}

	return ToEntity(&model), nil
}

func (r *Repository) Update(ctx context.Context, team *entity.Team) error {
	model := FromEntity(team)

	query := `
		UPDATE teams
		SET updated_at = $2
		WHERE team_name = $1
	`

	result, err := r.getDB(ctx).ExecContext(
		ctx,
		query,
		model.Name,
		model.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update team: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("team not found: %s", model.Name)
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, name string) error {
	query := `DELETE FROM teams WHERE team_name = $1`

	result, err := r.getDB(ctx).ExecContext(ctx, query, name)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("team not found: %s", name)
	}

	return nil
}

func (r *Repository) Exists(ctx context.Context, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`

	var exists bool
	err := r.getDB(ctx).QueryRowContext(ctx, query, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check team existence: %w", err)
	}

	return exists, nil
}
