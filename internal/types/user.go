package types

import "database/sql"

type User struct {
	ID           int64
	FirstName    string
	LastName     string
	Username     string
	LanguageCode string
	CreatedAt    sql.NullTime
	UpdatedAt    sql.NullTime
	DeletedAt    sql.NullTime
}

func (u *User) IsEmpty() bool {
	return u.ID == 0
}
