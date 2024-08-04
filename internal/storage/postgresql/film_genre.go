package postgresql

import (
	"context"
	"fmt"
	"whattowatch/internal/types"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (pg *PostgreSQL) InsertFilmGenre(ctx context.Context, genre types.FilmGenre) error {
	sql, args, err := sq.Insert("film_genres").Columns(
		"id", "tmdb_id", "name", "slug",
	).Values(genre.ID, genre.TMDbID, genre.Name, genre.Slug).PlaceholderFormat(sq.Dollar).Suffix("ON CONFLICT DO NOTHING").ToSql()
	if err != nil {
		return err
	}
	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert film genre: %s", err.Error())
	}
	return nil
}

func (pg *PostgreSQL) getGenreUUIDsByTMDBIDs(ctx context.Context, tmdbGenreIDs []int32) ([]uuid.UUID, error) {
	sql, args, err := sq.Select("id").From("film_genres").Where("tmdb_id = any(?)", tmdbGenreIDs).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := pg.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	genreUUIDs := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		genreUUIDs = append(genreUUIDs, id)
	}
	return genreUUIDs, nil
}

func (pg *PostgreSQL) InsertFilmContentGenres(ctx context.Context, filmContentID uuid.UUID, tmdbGenreIDs []int32) error {
	if len(tmdbGenreIDs) == 0 {
		return nil
	}

	_, err := pg.GetFilmContent(ctx, filmContentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		return err
	}

	genreUUIDs, err := pg.getGenreUUIDsByTMDBIDs(ctx, tmdbGenreIDs)
	if err != nil {
		return err
	}
	builder := sq.Insert("film_content_genres").Columns("film_content_id", "film_genre_id")
	for _, genreID := range genreUUIDs {
		builder = builder.Values(filmContentID, genreID)
	}
	sql, args, err := builder.PlaceholderFormat(sq.Dollar).Suffix("ON CONFLICT DO NOTHING").ToSql()
	if err != nil {
		return err
	}
	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert link between film content and genre: %s", err.Error())
	}
	return nil
}

func (pg *PostgreSQL) InsertFilmGenres(ctx context.Context, genres []types.FilmGenre) error {
	builder := sq.Insert("film_genres").Columns("id", "tmdb_id", "name", "slug")
	for _, genre := range genres {
		builder = builder.Values(genre.ID, genre.TMDbID, genre.Name, genre.Slug)
	}
	sql, args, err := builder.PlaceholderFormat(sq.Dollar).Suffix("ON CONFLICT DO NOTHING").ToSql()
	if err != nil {
		return err
	}
	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert film genre: %s", err.Error())
	}
	return nil
}
