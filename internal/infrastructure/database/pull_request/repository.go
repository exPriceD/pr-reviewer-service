package pull_request

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	trmsql "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"

	"github.com/exPriceD/pr-reviewer-service/internal/domain/entity"
	"github.com/exPriceD/pr-reviewer-service/internal/domain/repository"
)

var _ repository.PullRequestRepository = (*Repository)(nil)

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

func (r *Repository) Create(ctx context.Context, pr *entity.PullRequest) error {
	model := FromEntity(pr)

	query := `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at, merged_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.getDB(ctx).ExecContext(
		ctx,
		query,
		model.ID,
		model.Name,
		model.AuthorID,
		model.Status,
		model.CreatedAt,
		model.MergedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	if len(pr.AssignedReviewers()) > 0 {
		if err := r.insertReviewers(ctx, pr.ID(), pr.AssignedReviewers()); err != nil {
			return fmt.Errorf("failed to insert reviewers: %w", err)
		}
	}

	return nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*entity.PullRequest, error) {
	query := `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`

	var model Model
	err := r.getDB(ctx).QueryRowContext(ctx, query, id).Scan(
		&model.ID,
		&model.Name,
		&model.AuthorID,
		&model.Status,
		&model.CreatedAt,
		&model.MergedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find pull request: %w", err)
	}

	reviewers, err := r.findReviewersByPRID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find reviewers: %w", err)
	}

	return ToEntity(&model, reviewers), nil
}

// FindByIDForUpdate находит PR с блокировкой строки (SELECT FOR UPDATE)
// ВАЖНО: Должен вызываться внутри транзакции (txManager.Do)
func (r *Repository) FindByIDForUpdate(ctx context.Context, id string) (*entity.PullRequest, error) {
	query := `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
		FOR UPDATE
	`

	var model Model
	err := r.getDB(ctx).QueryRowContext(ctx, query, id).Scan(
		&model.ID,
		&model.Name,
		&model.AuthorID,
		&model.Status,
		&model.CreatedAt,
		&model.MergedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find pull request for update: %w", err)
	}

	reviewers, err := r.findReviewersByPRID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find reviewers: %w", err)
	}

	return ToEntity(&model, reviewers), nil
}

func (r *Repository) FindByReviewerID(ctx context.Context, reviewerID string) ([]*entity.PullRequest, error) {
	query := `
		SELECT DISTINCT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at
		FROM pull_requests pr
		INNER JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.user_id = $1
		ORDER BY pr.created_at DESC
	`

	rows, err := r.getDB(ctx).QueryContext(ctx, query, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to find pull requests by reviewer: %w", err)
	}
	//nolint:gosec
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	return r.scanPullRequestsFromRows(ctx, rows)
}

func (r *Repository) FindByAuthorID(ctx context.Context, authorID string) ([]*entity.PullRequest, error) {
	query := `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE author_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.getDB(ctx).QueryContext(ctx, query, authorID)
	if err != nil {
		return nil, fmt.Errorf("failed to find pull requests by author: %w", err)
	}
	//nolint:gosec
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	return r.scanPullRequestsFromRows(ctx, rows)
}

func (r *Repository) scanPullRequestsFromRows(ctx context.Context, rows *sql.Rows) ([]*entity.PullRequest, error) {
	var pullRequests []*entity.PullRequest
	for rows.Next() {
		var model Model
		if err := rows.Scan(&model.ID, &model.Name, &model.AuthorID, &model.Status, &model.CreatedAt, &model.MergedAt); err != nil {
			return nil, fmt.Errorf("failed to scan pull request: %w", err)
		}

		reviewers, err := r.findReviewersByPRID(ctx, model.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to find reviewers: %w", err)
		}

		pullRequests = append(pullRequests, ToEntity(&model, reviewers))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return pullRequests, nil
}

func (r *Repository) Update(ctx context.Context, pr *entity.PullRequest) error {
	model := FromEntity(pr)

	query := `
		UPDATE pull_requests
		SET pull_request_name = $2, author_id = $3, status = $4, merged_at = $5
		WHERE pull_request_id = $1
	`

	result, err := r.getDB(ctx).ExecContext(
		ctx,
		query,
		model.ID,
		model.Name,
		model.AuthorID,
		model.Status,
		model.MergedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update pull request: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("pull request not found: %s", model.ID)
	}

	if err := r.deleteReviewers(ctx, pr.ID()); err != nil {
		return fmt.Errorf("failed to delete reviewers: %w", err)
	}

	if len(pr.AssignedReviewers()) > 0 {
		if err := r.insertReviewers(ctx, pr.ID(), pr.AssignedReviewers()); err != nil {
			return fmt.Errorf("failed to insert reviewers: %w", err)
		}
	}

	return nil
}

// ReplaceReviewer заменяет одного ревьювера на другого одним запросом (оптимизация для ReassignReviewer)
func (r *Repository) ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	query := `
		UPDATE pr_reviewers
		SET user_id = $3
		WHERE pull_request_id = $1 AND user_id = $2
	`

	result, err := r.getDB(ctx).ExecContext(ctx, query, prID, oldReviewerID, newReviewerID)
	if err != nil {
		return fmt.Errorf("failed to replace reviewer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("reviewer not found or already replaced: pr_id=%s, old_reviewer_id=%s", prID, oldReviewerID)
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM pull_requests WHERE pull_request_id = $1`

	result, err := r.getDB(ctx).ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete pull request: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("pull request not found: %s", id)
	}

	return nil
}

// MergePR атомарно меняет статус PR на MERGED
func (r *Repository) MergePR(ctx context.Context, prID string) error {
	query := `
		WITH check_status AS (
			SELECT pull_request_id
			FROM pull_requests
			WHERE pull_request_id = $1 AND status = $2
		)
		UPDATE pull_requests
		SET 
			status = $3,
			merged_at = NOW()
		FROM check_status
		WHERE pull_requests.pull_request_id = check_status.pull_request_id
		RETURNING pull_requests.pull_request_id
	`

	var returnedID string
	err := r.getDB(ctx).QueryRowContext(ctx, query, prID, string(entity.PRStatusOpen), string(entity.PRStatusMerged)).Scan(&returnedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.ErrNotFound
		}
		return fmt.Errorf("failed to merge pull request: %w", err)
	}

	return nil
}

func (r *Repository) Exists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)`

	var exists bool
	err := r.getDB(ctx).QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check pull request existence: %w", err)
	}

	return exists, nil
}

func (r *Repository) findReviewersByPRID(ctx context.Context, prID string) ([]string, error) {
	query := `
		SELECT user_id
		FROM pr_reviewers
		WHERE pull_request_id = $1
		ORDER BY assigned_at
	`

	rows, err := r.getDB(ctx).QueryContext(ctx, query, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to find reviewers: %w", err)
	}
	//nolint:gosec
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var reviewers []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan reviewer: %w", err)
		}
		reviewers = append(reviewers, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return reviewers, nil
}

func (r *Repository) insertReviewers(ctx context.Context, prID string, reviewers []string) error {
	if len(reviewers) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(reviewers))
	valueArgs := make([]interface{}, 0, len(reviewers)*2)
	for i, reviewer := range reviewers {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, prID, reviewer)
	}

	query := fmt.Sprintf(`
		INSERT INTO pr_reviewers (pull_request_id, user_id)
		VALUES %s
	`, strings.Join(valueStrings, ","))

	_, err := r.getDB(ctx).ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to insert reviewers: %w", err)
	}

	return nil
}

func (r *Repository) deleteReviewers(ctx context.Context, prID string) error {
	query := `DELETE FROM pr_reviewers WHERE pull_request_id = $1`

	_, err := r.getDB(ctx).ExecContext(ctx, query, prID)
	if err != nil {
		return fmt.Errorf("failed to delete reviewers: %w", err)
	}

	return nil
}

func (r *Repository) CountActiveReviewsByUserIDs(ctx context.Context, userIDs []string) (map[string]int, error) {
	if len(userIDs) == 0 {
		return make(map[string]int), nil
	}

	placeholders := make([]string, len(userIDs))
	args := make([]interface{}, len(userIDs)+1)
	for i, userID := range userIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = userID
	}
	args[len(userIDs)] = string(entity.PRStatusOpen)

	query := fmt.Sprintf(`
		SELECT pr.user_id, COUNT(DISTINCT pr.pull_request_id) as review_count
		FROM pr_reviewers pr
		INNER JOIN pull_requests p ON pr.pull_request_id = p.pull_request_id
		WHERE pr.user_id IN (%s) AND p.status = $%d
		GROUP BY pr.user_id
	`, strings.Join(placeholders, ","), len(userIDs)+1)

	rows, err := r.getDB(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query active reviews count: %w", err)
	}
	//nolint:gosec
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	result := make(map[string]int)
	for _, userID := range userIDs {
		result[userID] = 0
	}

	for rows.Next() {
		var userID string
		var count int
		if err := rows.Scan(&userID, &count); err != nil {
			return nil, fmt.Errorf("failed to scan review count: %w", err)
		}
		result[userID] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return result, nil
}

// GetStats возвращает статистику по PR (общее количество, открытые, смерженные)
func (r *Repository) GetStats(ctx context.Context) (total, open, merged int, err error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = $1) as open,
			COUNT(*) FILTER (WHERE status = $2) as merged
		FROM pull_requests
	`

	err = r.getDB(ctx).QueryRowContext(ctx, query, string(entity.PRStatusOpen), string(entity.PRStatusMerged)).Scan(&total, &open, &merged)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get PR stats: %w", err)
	}

	return total, open, merged, nil
}

// CountReviewsByUserIDs возвращает общее количество назначений для каждого пользователя
// (включая как открытые, так и смерженные PR)
func (r *Repository) CountReviewsByUserIDs(ctx context.Context, userIDs []string) (map[string]int, error) {
	if len(userIDs) == 0 {
		return make(map[string]int), nil
	}

	placeholders := make([]string, len(userIDs))
	args := make([]interface{}, len(userIDs))
	for i, userID := range userIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = userID
	}

	query := fmt.Sprintf(`
		SELECT pr.user_id, COUNT(DISTINCT pr.pull_request_id) as review_count
		FROM pr_reviewers pr
		WHERE pr.user_id IN (%s)
		GROUP BY pr.user_id
	`, strings.Join(placeholders, ","))

	rows, err := r.getDB(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query reviews count: %w", err)
	}
	//nolint:gosec
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	result := make(map[string]int)
	for _, userID := range userIDs {
		result[userID] = 0
	}

	for rows.Next() {
		var userID string
		var count int
		if err := rows.Scan(&userID, &count); err != nil {
			return nil, fmt.Errorf("failed to scan review count: %w", err)
		}
		result[userID] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return result, nil
}
