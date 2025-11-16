package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	trmsql "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"

	"github.com/exPriceD/pr-reviewer-service/internal/domain/entity"
	"github.com/exPriceD/pr-reviewer-service/internal/domain/repository"
)

var _ repository.UserRepository = (*Repository)(nil)

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

func (r *Repository) Create(ctx context.Context, user *entity.User) error {
	model := FromEntity(user)

	query := `
		INSERT INTO users (user_id, username, team_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.getDB(ctx).ExecContext(
		ctx,
		query,
		model.ID,
		model.Username,
		model.TeamName,
		model.IsActive,
		model.CreatedAt,
		model.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
		SELECT user_id, username, team_name, is_active, created_at, updated_at
		FROM users
		WHERE user_id = $1
	`

	var model Model
	err := r.getDB(ctx).QueryRowContext(ctx, query, id).Scan(
		&model.ID,
		&model.Username,
		&model.TeamName,
		&model.IsActive,
		&model.CreatedAt,
		&model.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return ToEntity(&model), nil
}

func (r *Repository) FindByTeamName(ctx context.Context, teamName string) ([]*entity.User, error) {
	query := `
		SELECT user_id, username, team_name, is_active, created_at, updated_at
		FROM users
		WHERE team_name = $1
		ORDER BY username
	`

	rows, err := r.getDB(ctx).QueryContext(ctx, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to find users by team: %w", err)
	}
	//nolint:gosec
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	return r.scanUsersFromRows(rows)
}

func (r *Repository) FindActiveByTeamName(ctx context.Context, teamName string) ([]*entity.User, error) {
	query := `
		SELECT user_id, username, team_name, is_active, created_at, updated_at
		FROM users
		WHERE team_name = $1 AND is_active = true
		ORDER BY username
	`

	rows, err := r.getDB(ctx).QueryContext(ctx, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to find active users by team: %w", err)
	}
	//nolint:gosec
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	return r.scanUsersFromRows(rows)
}

func (r *Repository) scanUsersFromRows(rows *sql.Rows) ([]*entity.User, error) {
	var users []*entity.User
	for rows.Next() {
		var model Model
		if err := rows.Scan(&model.ID, &model.Username, &model.TeamName, &model.IsActive, &model.CreatedAt, &model.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, ToEntity(&model))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return users, nil
}

func (r *Repository) Update(ctx context.Context, user *entity.User) error {
	model := FromEntity(user)

	query := `
		UPDATE users
		SET username = $2, team_name = $3, is_active = $4, updated_at = $5
		WHERE user_id = $1
	`

	result, err := r.getDB(ctx).ExecContext(
		ctx,
		query,
		model.ID,
		model.Username,
		model.TeamName,
		model.IsActive,
		model.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %s", model.ID)
	}

	return nil
}

// BatchDeactivateByTeamName массово деактивирует всех активных пользователей команды одним запросом
func (r *Repository) BatchDeactivateByTeamName(ctx context.Context, teamName string) error {
	query := `
		UPDATE users
		SET is_active = FALSE, updated_at = NOW()
		WHERE team_name = $1 AND is_active = TRUE
	`

	_, err := r.getDB(ctx).ExecContext(ctx, query, teamName)
	if err != nil {
		return fmt.Errorf("failed to batch deactivate team members: %w", err)
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE user_id = $1`

	result, err := r.getDB(ctx).ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %s", id)
	}

	return nil
}

func (r *Repository) Exists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)`

	var exists bool
	err := r.getDB(ctx).QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}
