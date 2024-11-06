package postgresql

import (
	"context"
	"fmt"
	"whattowatch/internal/types"

	sq "github.com/Masterminds/squirrel"
)

func (pg *PostgreSQL) GetContentItem(ctx context.Context, id int64) (types.ContentItem, error) {
	sql, args, err := sq.Select("*").PlaceholderFormat(sq.Dollar).From("content").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return types.ContentItem{}, err
	}
	var fc types.ContentItem
	err = pg.conn.QueryRow(ctx, sql, args...).Scan(
		&fc.ID,
		&fc.ContentType,
		&fc.Title,
		&fc.Overview,
		&fc.Popularity,
		&fc.PosterPath,
		&fc.ReleaseDate,
		&fc.VoteAverage,
		&fc.VoteCount,
	)
	if err != nil {
		return types.ContentItem{}, err
	}

	fc.Genres, err = pg.GetGenres(ctx, fc.ID)
	if err != nil {
		return types.ContentItem{}, err
	}

	return fc, nil
}

func (pg *PostgreSQL) InsertContent(ctx context.Context, content types.Content) error {
	contentBuilder := sq.Insert("content").Columns(
		"id",
		"content_type_id",
		"title",
		"overview",
		"popularity",
		"poster_path",
		"release_date",
		"vote_average",
		"vote_count",
	).PlaceholderFormat(sq.Dollar)

	genresBuilder := sq.Insert("link_content_genres").
		Columns("content_id", "genre_id").
		PlaceholderFormat(sq.Dollar)

	for _, c := range content {
		contentBuilder = contentBuilder.Values(
			c.ID,
			c.ContentType.EnumIndex(),
			c.Title,
			c.Overview,
			c.Popularity,
			c.PosterPath,
			c.ReleaseDate,
			c.VoteAverage,
			c.VoteCount,
		)

		for _, g := range c.Genres {
			genresBuilder = genresBuilder.Values(c.ID, g.ID)
		}
	}

	contentSql, contentArgs, err := contentBuilder.Suffix("ON CONFLICT DO NOTHING").ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert content query: %s", err.Error())
	}

	genresSql, genresArgs, err := genresBuilder.Suffix("ON CONFLICT DO NOTHING").ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert content genres query: %s", err.Error())
	}

	tx, err := pg.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %s", err.Error())
	}
	defer tx.Rollback(ctx)

	// Insert content and genres
	_, err = tx.Exec(ctx, contentSql, contentArgs...)
	if err != nil {
		return fmt.Errorf("failed to insert content: %s; data: %v", err.Error(), content)
	}
	// pg.log.Debug("content inserted", "type", content[0].ContentType, "ids", content.GetIDsWithGenres())

	_, err = tx.Exec(ctx, genresSql, genresArgs...)
	if err != nil {
		return fmt.Errorf("failed to insert content genres: %s; ids: %v; kv-ids: %v", err.Error(), content.GetIDs(), content.GetIDsWithGenres())
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %s", err.Error())
	}

	return nil
}

func (pg *PostgreSQL) GetContentStatus(ctx context.Context, userID int64, contentID int64) (types.ContentStatus, error) {
	favoriteSQL, favArgs, err := sq.Select("*").
		From("users_favorites t1").
		Where(sq.Eq{"t1.user_id": userID, "t1.content_id": contentID}).ToSql()

	if err != nil {
		return types.ContentStatus{}, fmt.Errorf("failed to build favorite subquery: %s", err.Error())
	}

	viewedSQL, viewArgs, err := sq.Select("*").
		From("users_viewed t2").
		Where(sq.Eq{"t2.user_id": userID, "t2.content_id": contentID}).ToSql()

	if err != nil {
		return types.ContentStatus{}, fmt.Errorf("failed to build viewed subquery: %s", err.Error())
	}

	query := sq.Select(
		fmt.Sprintf("EXISTS(%s) AS is_favorite", favoriteSQL),
		fmt.Sprintf("EXISTS(%s) AS is_viewed", viewedSQL),
	).PlaceholderFormat(sq.Dollar)

	sql, _, err := query.ToSql()
	if err != nil {
		return types.ContentStatus{}, fmt.Errorf("failed to build main query: %s", err.Error())
	}

	cs := types.ContentStatus{
		UserID:    userID,
		ContentID: contentID,
	}

	args := append(favArgs, viewArgs...)

	err = pg.conn.QueryRow(ctx, sql, args...).Scan(&cs.IsFavorite, &cs.IsViewed)
	if err != nil {
		pg.log.Error("failed to get content status", "error", err.Error(), "sql", sql, "args", args)
		return types.ContentStatus{}, fmt.Errorf("failed to get content: %s", err.Error())
	}

	return cs, nil
}

func (pg *PostgreSQL) AddContentItemToFavorite(ctx context.Context, userID int64, contentID int64) error {
	sql, args, err := sq.Insert("users_favorites").Columns("user_id", "content_id").Values(userID, contentID).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build sql query: %s", err.Error())
	}

	_, err = pg.conn.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert favorite: %s", err.Error())
	}
	return nil
}

func (pg *PostgreSQL) RemoveContentItemFromFavorite(ctx context.Context, userID int64, contentID int64) error {
	sql, args, err := sq.Delete("users_favorites").Where(sq.Eq{"user_id": userID, "content_id": contentID}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build sql query: %s", err.Error())
	}

	_, err = pg.conn.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert favorite: %s", err.Error())
	}
	return nil
}

func (pg *PostgreSQL) GetFavoriteContent(ctx context.Context, userID int64) (types.Content, error) {
	sql, args, err := sq.Select("t1.*").
		From("content t1").
		Join("users_favorites t2 ON t1.id = t2.content_id").
		Where(sq.Eq{"t2.user_id": userID}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql query: %s", err.Error())
	}

	rows, err := pg.conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get favorite content: %s", err.Error())
	}
	defer rows.Close()

	content := make(types.Content, 0)
	for rows.Next() {
		c := types.ContentItem{}
		err = rows.Scan(
			&c.ID,
			&c.ContentType,
			&c.Title,
			&c.Overview,
			&c.Popularity,
			&c.PosterPath,
			&c.ReleaseDate,
			&c.VoteAverage,
			&c.VoteCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %s", err.Error())
		}
		content = append(content, c)
	}
	return content, nil
}

func (pg *PostgreSQL) AddContentItemToViewed(ctx context.Context, userID int64, contentID int64) error {
	sql, args, err := sq.Insert("users_viewed").Columns("user_id", "content_id").Values(userID, contentID).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build sql query: %s", err.Error())
	}

	_, err = pg.conn.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert viewed: %s", err.Error())
	}
	return nil
}

func (pg *PostgreSQL) RemoveContentItemFromViewed(ctx context.Context, userID int64, contentID int64) error {
	sql, args, err := sq.Delete("users_viewed").Where(sq.Eq{"user_id": userID, "content_id": contentID}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build sql query: %s", err.Error())
	}

	_, err = pg.conn.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert viewed: %s", err.Error())
	}
	return nil
}

func (pg *PostgreSQL) GetViewedContent(ctx context.Context, userID int64) (types.Content, error) {
	sql, args, err := sq.Select("t1.*").
		From("content t1").
		Join("users_viewed t2 ON t1.id = t2.content_id").
		Where(sq.Eq{"t2.user_id": userID}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql query: %s", err.Error())
	}

	rows, err := pg.conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get viewed content: %s", err.Error())
	}
	defer rows.Close()

	content := make(types.Content, 0)
	for rows.Next() {
		c := types.ContentItem{}
		err = rows.Scan(
			&c.ID,
			&c.ContentType,
			&c.Title,
			&c.Overview,
			&c.Popularity,
			&c.PosterPath,
			&c.ReleaseDate,
			&c.VoteAverage,
			&c.VoteCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %s", err.Error())
		}
		content = append(content, c)
	}
	return content, nil
}
