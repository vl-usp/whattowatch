package postgresql

import (
	"context"
	"fmt"
	"whattowatch/internal/types"

	sq "github.com/Masterminds/squirrel"
)

func (pg *PostgreSQL) GetContent(ctx context.Context, id int64) (types.Content, error) {
	sql, args, err := sq.Select("*").PlaceholderFormat(sq.Dollar).From("content").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return types.Content{}, err
	}
	var fc types.Content
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
		return types.Content{}, err
	}
	return fc, nil
}

func (pg *PostgreSQL) InsertContentSlice(ctx context.Context, contents types.ContentSlice) error {
	builder := sq.Insert("content").Columns(
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

	for _, c := range contents {
		builder = builder.Values(
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
	}

	sql, args, err := builder.Suffix("ON CONFLICT DO NOTHING").ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert content query: %s", err.Error())
	}

	_, err = pg.conn.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert content: %s", err.Error())
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

func (pg *PostgreSQL) AddContentToFavorite(ctx context.Context, userID int64, contentID int64) error {
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

func (pg *PostgreSQL) RemoveContentFromFavorite(ctx context.Context, userID int64, contentID int64) error {
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

func (pg *PostgreSQL) AddContentToViewed(ctx context.Context, userID int64, contentID int64) error {
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

func (pg *PostgreSQL) RemoveContentFromViewed(ctx context.Context, userID int64, contentID int64) error {
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
