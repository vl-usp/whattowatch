package postgresql

import (
	"context"
	"fmt"
	"whattowatch/internal/types"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid"
)

func (pg *PostgreSQL) GetContent(ctx context.Context, id uuid.UUID) (types.Content, error) {
	sql, args, err := sq.Select("*").PlaceholderFormat(sq.Dollar).From("content").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return types.Content{}, err
	}
	var fc types.Content
	err = pg.pool.QueryRow(ctx, sql, args...).Scan(
		&fc.ID,
		&fc.TMDbID,
		&fc.ContentTypeID,
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

func (pg *PostgreSQL) GetContentTMDbIDs(ctx context.Context) ([]uuid.UUID, error) {
	sql, args, err := sq.Select("tmdb_id").PlaceholderFormat(sq.Dollar).From("content").ToSql()
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0)
	r, err := pg.pool.Query(ctx, sql, args...)
	if err != nil {
		return ids, fmt.Errorf("failed to get ids: %s", err.Error())
	}
	for r.Next() {
		var id uuid.UUID
		err = r.Scan(&id)
		if err != nil {
			return ids, fmt.Errorf("failed to scan: %s", err.Error())
		}
		ids = append(ids, id)
	}
	r.Close()
	return ids, nil
}

func (pg *PostgreSQL) GetContentByTitles(ctx context.Context, titles []string) (types.Contents, error) {
	builder := sq.Select("*").PlaceholderFormat(sq.Dollar).From("content")
	sql, args, err := builder.Where("title = any(?)", titles).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := pg.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	contentData := make(types.Contents, 0, len(rows.RawValues()))

	for rows.Next() {
		var content types.Content
		err = rows.Scan(
			&content.ID,
			&content.TMDbID,
			&content.ContentTypeID,
			&content.Title,
			&content.Overview,
			&content.Popularity,
			&content.PosterPath,
			&content.ReleaseDate,
			&content.VoteAverage,
			&content.VoteCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan movie from db: %s", err.Error())
		}
		contentData = append(contentData, content)
	}

	return contentData, nil
}

func (pg *PostgreSQL) InsertContent(ctx context.Context, content types.Content) error {
	sql1, args1, err := sq.Insert("content").SetMap(sq.Eq{
		"id":              content.ID,
		"content_type_id": content.ContentTypeID,
		"title":           content.Title,
		"overview":        content.Overview,
		"popularity":      content.Popularity,
		"poster_path":     content.PosterPath,
		"release_date":    content.ReleaseDate,
		"vote_average":    content.VoteAverage,
		"vote_count":      content.VoteCount,
	}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %s", err.Error())
	}

	sql2, args2, err := sq.Insert("link_tmdb_contents").SetMap(sq.Eq{
		"content_id": content.ID,
		"tmdb_id":    content.TMDbID,
	}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %s", err.Error())
	}

	tx, err := pg.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = pg.pool.Exec(ctx, sql1, args1...)
	if err != nil {
		return fmt.Errorf("failed to insert content: %s, %v", err.Error(), content)
	}

	_, err = pg.pool.Exec(ctx, sql2, args2...)
	if err != nil {
		return fmt.Errorf("failed to insert content: %s, %v", err.Error(), content)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (pg *PostgreSQL) InsertContents(ctx context.Context, contents types.Contents) error {
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
			c.ContentTypeID,
			c.Title,
			c.Overview,
			c.Popularity,
			c.PosterPath,
			c.ReleaseDate,
			c.VoteAverage,
			c.VoteCount,
		)
	}

	sql1, args1, err := builder.Suffix("ON CONFLICT DO NOTHING").ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %s", err.Error())
	}

	builder = sq.Insert("link_tmdb_contents").Columns("content_id", "tmdb_id")
	for _, c := range contents {
		builder = builder.Values(c.ID, c.TMDbID)
	}
	sql2, args2, err := builder.Suffix("ON CONFLICT DO NOTHING").ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %s", err.Error())
	}

	tx, err := pg.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = pg.pool.Exec(ctx, sql1, args1...)
	if err != nil {
		return fmt.Errorf("failed to insert film content: %s", err.Error())
	}

	_, err = pg.pool.Exec(ctx, sql2, args2...)
	if err != nil {
		return fmt.Errorf("failed to insert film content: %s", err.Error())
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (pg *PostgreSQL) UpdateContent(ctx context.Context, movie types.Content) error {
	sql, args, err := sq.Update("content").SetMap(sq.Eq{
		"popularity":   movie.Popularity,
		"poster_path":  movie.PosterPath,
		"release_date": movie.ReleaseDate,
		"vote_average": movie.VoteAverage,
		"vote_count":   movie.VoteCount,
	}).Where(sq.Eq{"id": movie.ID}).PlaceholderFormat(sq.Dollar).ToSql()

	if err != nil {
		return fmt.Errorf("failed to build update query: %s", err.Error())
	}
	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to update tmdb content: %s", err.Error())
	}
	return nil
}
