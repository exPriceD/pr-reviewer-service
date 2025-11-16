package pull_request

import (
	"database/sql"
	"time"
)

type Model struct {
	ID        string       `db:"pull_request_id"`
	Name      string       `db:"pull_request_name"`
	AuthorID  string       `db:"author_id"`
	Status    string       `db:"status"`
	CreatedAt time.Time    `db:"created_at"`
	MergedAt  sql.NullTime `db:"merged_at"`
}
