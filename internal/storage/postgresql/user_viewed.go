package postgresql

import (
	"context"
	"fmt"
	"whattowatch/internal/types"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid"
)

func (pg *PostgreSQL) InsertUserViewed(ctx context.Context, userID int, filmContentID uuid.UUID) error {
	builder := sq.Insert("user_viewed").Columns("user_id", "content_id")
	sql, args, err := builder.Values(userID, filmContentID).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build sql script for inserting user viewed. error: %s, user_id: %d, content_id: %s", err.Error(), userID, filmContentID.String())
	}

	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert user viewed. error: %s, user_id: %d, content_id: %s", err.Error(), userID, filmContentID.String())
	}
	return nil
}

func (pg *PostgreSQL) InsertUserVieweds(ctx context.Context, userID int, filmContentIDs []uuid.UUID) error {
	builder := sq.Insert("user_viewed").Columns("user_id", "content_id")
	for _, id := range filmContentIDs {
		builder = builder.Values(userID, id)
	}
	sql, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build sql script for inserting user viewed. error: %s, user_id: %d", err.Error(), userID)
	}

	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert user viewed. error: %s, user_id: %d", err.Error(), userID)
	}
	return nil
}

func (pg *PostgreSQL) GetUserViewed(ctx context.Context, userID int) (types.UserVieweds, error) {
	sql, args, err := sq.Select("*").From("user_viewed").PlaceholderFormat(sq.Dollar).Where(sq.Eq{"user_id": userID}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql script for getting user viewed. error: %s, user_id: %d", err.Error(), userID)
	}
	rows, err := pg.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get tg user viewed: %s", err.Error())
	}
	defer rows.Close()
	vieweds := make(types.UserVieweds, 0)
	for rows.Next() {
		var userViewed types.UserViewed
		err = rows.Scan(&userViewed.ID, &userViewed.UserID, &userViewed.ContentID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user viewed fIDm db: %s", err.Error())
		}
		vieweds = append(vieweds, userViewed)
	}
	return vieweds, nil
}
