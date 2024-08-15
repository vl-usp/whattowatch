package postgresql

import (
	"context"
	"fmt"
	"sort"
	"whattowatch/internal/types"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid"
)

func (pg *PostgreSQL) GetUserFavorites(ctx context.Context, userID int) (types.Contents, error) {
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
		err = rows.Scan(&userFavorite.ID, &userFavorite.UserID, &userFavorite.FavoriteID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user favorite fIDm db: %s", err.Error())
		}
		favorites = append(favorites, userFavorite)
	}

	contentData := make(types.Contents, 0)
	sql, args, err = sq.Select("*").From("content").PlaceholderFormat(sq.Dollar).Where("id = any(?)", favorites.FavoriteIDs()).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql script for getting film content error: %s", err.Error())
	}
	rows, err = pg.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get content film content favorites: %s", err.Error())
	}
	for rows.Next() {
		var fc types.Content
		err = rows.Scan(&fc.ID, &fc.TMDbID, &fc.ContentTypeID, &fc.Title, &fc.Overview, &fc.Popularity, &fc.PosterPath, &fc.ReleaseDate, &fc.VoteAverage, &fc.VoteCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan film content from db: %s", err.Error())
		}

		genres, err := pg.GetContentGenres(ctx, fc.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get film content genres: %s", err.Error())
		}
		fc.Genres = genres
		contentData = append(contentData, fc)
	}

	return contentData, nil
}

func (pg *PostgreSQL) GetUserFavoritesIDs(ctx context.Context, userID int) ([]uuid.UUID, error) {
	sql, args, err := sq.Select("favorite_id").From("user_favorites").PlaceholderFormat(sq.Dollar).Where(sq.Eq{"user_id": userID}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql script for getting user favorite id error: %s, user_id: %d", err.Error(), userID)
	}
	rows, err := pg.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get tg user favorite: %s", err.Error())
	}
	defer rows.Close()
	favoriteIDs := make([]uuid.UUID, 0, len(rows.RawValues()))
	for rows.Next() {
		var id uuid.UUID
		err = rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user favorite fIDm db: %s", err.Error())
		}
		favoriteIDs = append(favoriteIDs, id)
	}
	return favoriteIDs, nil
}

func (pg *PostgreSQL) GetUserFavoriteIDByTitle(ctx context.Context, userID int, title string) (uuid.UUID, error) {
	sql, args, err := sq.Select("favorite_id").From("user_favorites t1").
		Join("content t2 on t2.id = t1.favorite_id").
		PlaceholderFormat(sq.Dollar).Where("t1.user_id = ? and t2.title = ?", userID, title).ToSql()
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to build sql script for getting user favorite id error: %s, user_id: %d", err.Error(), userID)
	}
	var id uuid.UUID
	err = pg.pool.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get tg user favorite: %s", err.Error())
	}
	return id, nil
}

func (pg *PostgreSQL) GetUserFavoritesByType(ctx context.Context, userID int) (types.ContentsByTypes, error) {
	data, err := pg.GetUserFavorites(ctx, userID)
	if err != nil {
		return nil, err
	}
	m := make(types.ContentsByTypes)
	for _, fc := range data {
		t := types.ContentType(fc.ContentTypeID)
		m[t] = append(m[t], fc)
	}

	for _, content := range m {
		sort.Slice(content, func(i, j int) bool {
			return content[i].Popularity < content[j].Popularity
		})
	}
	return m, nil
}

func (pg *PostgreSQL) InsertUserFavorites(ctx context.Context, userID int, filmContentIDs []uuid.UUID) error {
	// insert user favorites
	builder := sq.Insert("user_favorites").Columns("user_id", "favorite_id")
	for _, id := range filmContentIDs {
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

func (pg *PostgreSQL) DeleteUserFavorites(ctx context.Context, userID int, filmContentIDs []uuid.UUID) error {
	builder := sq.Delete("user_favorites").Where("user_id = ? AND favorite_id = any(?)", userID, filmContentIDs)
	sql, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build sql script for setting user favorite id error: %s", err.Error())
	}
	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete user favorite: %s", err.Error())
	}
	return nil
}
