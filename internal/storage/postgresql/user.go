package postgresql

import (
	"context"
	"fmt"
	"whattowatch/internal/types"

	sq "github.com/Masterminds/squirrel"
)

func (pg *PostgreSQL) GetUser(ctx context.Context, id int) (types.User, error) {
	sql, args, err := sq.Select("*").From("users").PlaceholderFormat(sq.Dollar).Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return types.User{}, err
	}
	var user types.User
	err = pg.conn.QueryRow(ctx, sql, args...).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.LanguageCode)
	if err != nil {
		return types.User{}, err
	}
	return user, nil
}

func (pg *PostgreSQL) InsertUser(ctx context.Context, user types.User) error {
	builder := sq.Insert("users").Columns("id", "first_name", "last_name", "username", "language_code", "created_at")
	builder = builder.Values(user.ID, user.FirstName, user.LastName, user.Username, user.LanguageCode, user.CreatedAt)
	sql, args, err := builder.Suffix("ON CONFLICT DO NOTHING").PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return err
	}
	_, err = pg.conn.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert tg user: %s", err.Error())
	}
	return nil
}
