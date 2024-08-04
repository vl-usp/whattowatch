package postgresql

import (
	"context"
	"fmt"
	"whattowatch/internal/types"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (pg *PostgreSQL) InsertUserFavorites(ctx context.Context, userID int, filmContentIds []uuid.UUID) error {
	// insert user favorites
	builder := sq.Insert("user_favorites").Columns("user_id", "favorite_id")
	for _, id := range filmContentIds {
		builder = builder.Values(userID, id)
	}
	sql, args, err := builder.Suffix("ON CONFLICT DO NOTHING").PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build sql script for setting user favorite id error: %s", err.Error())
	}
	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert user favorite: %s", err.Error())
	}
	return nil
}

func (pg *PostgreSQL) GetUserFavorites(ctx context.Context, userID int) (types.FilmContents, error) {
	sql, args, err := sq.Select("*").From("user_favorites").PlaceholderFormat(sq.Dollar).Where(sq.Eq{"user_id": userID}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql script for getting user favorite id error: %s, user_id: %d", err.Error(), userID)
	}
	rows, err := pg.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get tg user favorite: %s", err.Error())
	}
	defer rows.Close()
	favorites := make(types.UserFavorites, 0)

	for rows.Next() {
		var userFavorite types.UserFavorite
		err = rows.Scan(&userFavorite.ID, &userFavorite.UserID, &userFavorite.FilmContentType, &userFavorite.FavoriteID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user favorite from db: %s", err.Error())
		}
		favorites = append(favorites, userFavorite)
	}

	contentData := make(types.FilmContents, 0, len(favorites))
	sql, args, err = sq.Select("*").From("film_content").PlaceholderFormat(sq.Dollar).Where("id = any(?)", favorites).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql script for getting film content error: %s", err.Error())
	}
	rows, err = pg.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get content film content favorites: %s", err.Error())
	}
	for rows.Next() {
		var fc types.FilmContent
		err = rows.Scan(&fc.ID, &fc.TMDbID, &fc.FilmContentTypeId, &fc.Title, &fc.Overview, &fc.Popularity, &fc.PosterPath, &fc.ReleaseDate, &fc.VoteAverage, &fc.VoteCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan film content from db: %s", err.Error())
		}
		contentData = append(contentData, fc)
	}

	return contentData, nil
}
