package user

import "time"

type Model struct {
	ID        string    `db:"user_id"`
	Username  string    `db:"username"`
	TeamName  string    `db:"team_name"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
