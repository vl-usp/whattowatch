package types

import (
	"database/sql"
	"time"
)

type Source struct {
	ID        int
	Name      string
	Url       string
	CreatedAt time.Time
	UpdatedAt sql.NullTime
	DeletedAt sql.NullTime
}
