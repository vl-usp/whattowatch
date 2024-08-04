package postgresql

import (
	"context"
	"fmt"
	"whattowatch/internal/types"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (pg *PostgreSQL) GetFilmContent(ctx context.Context, id uuid.UUID) (types.FilmContent, error) {
	sql, args, err := sq.Select("*").PlaceholderFormat(sq.Dollar).From("film_content").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return types.FilmContent{}, err
	}
	var fc types.FilmContent
	err = pg.pool.QueryRow(ctx, sql, args...).Scan(
		&fc.ID,
		&fc.TMDbID,
		&fc.FilmContentTypeId,
		&fc.Title,
		&fc.Overview,
		&fc.Popularity,
		&fc.PosterPath,
		&fc.ReleaseDate,
		&fc.VoteAverage,
		&fc.VoteCount,
	)
	if err != nil {
		return types.FilmContent{}, err
	}
	return fc, nil
}

func (pg *PostgreSQL) GetFilmContentTMDbIDs(ctx context.Context) ([]uuid.UUID, error) {
	sql, args, err := sq.Select("tmdb_id").PlaceholderFormat(sq.Dollar).From("film_content").ToSql()
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

func (pg *PostgreSQL) GetFilmContentByTitles(ctx context.Context, titles []string) (types.FilmContents, error) {
	builder := sq.Select("*").PlaceholderFormat(sq.Dollar).From("film_content")
	sql, args, err := builder.Where("title = any(?)", titles).ToSql()
	if err != nil {
		return nil, err
	}
	fmt.Println(sql, args)

	rows, err := pg.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	contentData := make(types.FilmContents, 0, len(rows.RawValues()))

	for rows.Next() {
		var content types.FilmContent
		err = rows.Scan(
			&content.ID,
			&content.TMDbID,
			&content.FilmContentTypeId,
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
		content.FilmContentTypeId = 1
		fmt.Println(content)
		contentData = append(contentData, content)
	}

	return contentData, nil
}

func (pg *PostgreSQL) UpdateFilmContent(ctx context.Context, movie types.FilmContent) error {
	sql, args, err := sq.Update("film_content").SetMap(sq.Eq{
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

func (pg *PostgreSQL) InsertFilmContent(ctx context.Context, content types.FilmContent) error {
	sql, args, err := sq.Insert("film_content").SetMap(sq.Eq{
		"id":                   content.ID,
		"tmdb_id":              content.TMDbID,
		"film_content_type_id": content.FilmContentTypeId,
		"title":                content.Title,
		"overview":             content.Overview,
		"popularity":           content.Popularity,
		"poster_path":          content.PosterPath,
		"release_date":         content.ReleaseDate,
		"vote_average":         content.VoteAverage,
		"vote_count":           content.VoteCount,
	}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %s", err.Error())
	}

	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert film content: %s, %v", err.Error(), content)
	}
	return nil
}

func (pg *PostgreSQL) InsertFilmContents(ctx context.Context, contents types.FilmContents) error {
	builder := sq.Insert("film_content").Columns(
		"id",
		"tmdb_id",
		"film_content_type_id",
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
			c.TMDbID,
			c.FilmContentTypeId,
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
		return fmt.Errorf("failed to build insert query: %s", err.Error())
	}

	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert film content: %s", err.Error())
	}

	return nil
}
