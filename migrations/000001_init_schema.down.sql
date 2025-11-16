
DROP TABLE IF EXISTS pr_reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;

DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TRIGGER IF EXISTS update_teams_updated_at ON teams;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

DROP INDEX IF EXISTS idx_users_team_name;
DROP INDEX IF EXISTS idx_users_team_active;
DROP INDEX IF EXISTS idx_pr_reviewers_user;
DROP INDEX IF EXISTS idx_pr_reviewers_pr;
DROP INDEX IF EXISTS idx_pr_author;
DROP INDEX IF EXISTS idx_pr_created_at;
DROP INDEX IF EXISTS idx_pr_status;
