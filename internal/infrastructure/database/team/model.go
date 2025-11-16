package team

import "time"

type Model struct {
	Name      string    `db:"team_name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
