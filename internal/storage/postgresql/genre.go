package postgresql

import (
	"context"
	"fmt"
	"whattowatch/internal/types"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
)

func (pg *PostgreSQL) GetContentGenres(ctx context.Context, filmContentID uuid.UUID) (types.Genres, error) {
	sql, args, err := sq.Select("genres.id", "genres.tmdb_id", "genres.name", "genres.slug", "genres.formatted_name").
		From("genres").
		Join("content_genres ON genres.id = content_genres.genre_id").
		Where("content_genres.content_id = ?", filmContentID).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := pg.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	genres := make(types.Genres, 0)
	for rows.Next() {
		genre := types.Genre{}
		err = rows.Scan(&genre.ID, &genre.TMDbID, &genre.Name, &genre.Slug, &genre.FormattedName)
		if err != nil {
			return nil, err
		}
		genres = append(genres, genre)
	}
	return genres, nil
}

func (pg *PostgreSQL) GetGenresByIDs(ctx context.Context, ids []int) (types.Genres, error) {
	sql, args, err := sq.Select("id", "tmdb_id", "name", "slug", "formatted_name").
		From("genres").
		Where("tmdb_id = any(?)", ids).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := pg.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	genres := make(types.Genres, 0)
	for rows.Next() {
		genre := types.Genre{}
		err = rows.Scan(&genre.ID, &genre.TMDbID, &genre.Name, &genre.Slug, &genre.FormattedName)
		if err != nil {
			return nil, err
		}
		genres = append(genres, genre)
	}
	return genres, nil
}

func (pg *PostgreSQL) InsertGenre(ctx context.Context, genre types.Genre) error {
	sql, args, err := sq.Insert("genres").Columns(
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

func (pg *PostgreSQL) InsertContentGenres(ctx context.Context, filmContentID uuid.UUID, tmdbGenreIDs []int32) error {
	if len(tmdbGenreIDs) == 0 {
		return nil
	}

	_, err := pg.GetContent(ctx, filmContentID)
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
	builder := sq.Insert("content_genres").Columns("content_id", "genre_id")
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

func (pg *PostgreSQL) InsertGenres(ctx context.Context, genres types.Genres) error {
	builder := sq.Insert("genres").Columns("id", "tmdb_id", "name", "slug")
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

func (pg *PostgreSQL) getGenreUUIDsByTMDBIDs(ctx context.Context, tmdbGenreIDs []int32) ([]uuid.UUID, error) {
	sql, args, err := sq.Select("id").From("genres").Where("tmdb_id = any(?)", tmdbGenreIDs).PlaceholderFormat(sq.Dollar).ToSql()
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
