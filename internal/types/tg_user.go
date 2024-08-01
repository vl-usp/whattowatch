package types

import "database/sql"

type TGUser struct {
	ID           int64
	FirstName    string
	LastName     string
	Username     string
	LanguageCode string
	CreatedAt    sql.NullTime
	UpdatedAt    sql.NullTime
	DeletedAt    sql.NullTime
}
