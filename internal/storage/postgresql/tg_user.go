package postgresql

import (
	"context"
	"fmt"
	"whattowatch/internal/types"

	sq "github.com/Masterminds/squirrel"
)

func (pg *PostgreSQL) InsertTGUser(ctx context.Context, user types.TGUser) error {
	builder := sq.Insert("tg.users").Columns("id", "first_name", "last_name", "username", "language_code")
	builder = builder.Values(user.ID, user.FirstName, user.LastName, user.Username, user.LanguageCode)
	sql, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return err
	}
	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert tg user: %s", err.Error())
	}
	return nil
}
