-- Таблица команд
CREATE TABLE IF NOT EXISTS teams (
    team_name VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    team_name VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_users_team FOREIGN KEY (team_name) REFERENCES teams(team_name) ON DELETE RESTRICT ON UPDATE CASCADE
);

-- Справочник статусов PR
CREATE TABLE IF NOT EXISTS pr_statuses (
    status VARCHAR(50) PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Заполнение справочника статусов
INSERT INTO pr_statuses (status) VALUES ('OPEN'), ('MERGED') ON CONFLICT (status) DO NOTHING;

-- Таблица Pull Requests
CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id VARCHAR(255) PRIMARY KEY,
    pull_request_name VARCHAR(255) NOT NULL,
    author_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMPTZ,
    CONSTRAINT fk_pr_author FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_pr_status FOREIGN KEY (status) REFERENCES pr_statuses(status) ON DELETE RESTRICT ON UPDATE RESTRICT
);

-- Таблица связи PR и ревьюверов (m-m)
CREATE TABLE IF NOT EXISTS pr_reviewers (
    pull_request_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (pull_request_id, user_id),
    CONSTRAINT fk_pr_reviewers_pr FOREIGN KEY (pull_request_id) REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_pr_reviewers_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE RESTRICT ON UPDATE CASCADE
);

-- Для получения команды с участниками (/team/get)
CREATE INDEX idx_users_team_name ON users(team_name);

-- Для выборки активных пользователей команды при назначении ревьюверов
CREATE INDEX idx_users_team_active ON users(team_name) 
    WHERE is_active = TRUE;

-- Для получения PR где пользователь ревьювер (/users/getReview)
CREATE INDEX idx_pr_reviewers_user ON pr_reviewers(user_id);

-- Для получения ревьюверов конкретного PR
CREATE INDEX idx_pr_reviewers_pr ON pr_reviewers(pull_request_id);

-- Для статистики и сортировки
CREATE INDEX idx_pr_author ON pull_requests(author_id);
CREATE INDEX idx_pr_created_at ON pull_requests(created_at DESC);
CREATE INDEX idx_pr_status ON pull_requests(status);

-- Триггер для автоматического обновления updated_at для teams
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_teams_updated_at
    BEFORE UPDATE ON teams
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_pr_statuses_updated_at
    BEFORE UPDATE ON pr_statuses
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();