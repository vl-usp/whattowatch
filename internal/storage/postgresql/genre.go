package postgresql

import (
	"context"
	"fmt"
	"strings"
	"whattowatch/internal/types"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid"
)

func (pg *PostgreSQL) GetGenres(ctx context.Context, contentID uuid.UUID) (types.Genres, error) {
	sql, args, err := sq.Select("t1.id", "t2.tmdb_genre_id", "t1.name", "t1.pretty_name").
		From("genres t1").
		Join("link_tmdb_genres t2 ON t1.id = t2.genre_id").
		Join("link_content_genres t3 on t1.id = t3.genre_id").
		Where("t3.content_id = ?", contentID).PlaceholderFormat(sq.Dollar).ToSql()
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
		err = rows.Scan(&genre.ID, &genre.TMDbID, &genre.Name, &genre.PrettyName)
		if err != nil {
			return nil, err
		}
		genres = append(genres, genre)
	}

	return genres, nil
}

func (pg *PostgreSQL) GetGenresByIDs(ctx context.Context, ids []int) (types.Genres, error) {
	sql, args, err := sq.Select("t1.id", "t2.tmdb_genre_id", "t1.name", "t1.pretty_name").
		From("genres t1").
		Join("link_tmdb_genres t2 on t1.id = t2.genre_id").
		Where("t2.tmdb_genre_id = any(?)", ids).PlaceholderFormat(sq.Dollar).ToSql()
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
		err = rows.Scan(&genre.ID, &genre.TMDbID, &genre.Name, &genre.PrettyName)
		if err != nil {
			return nil, err
		}
		genres = append(genres, genre)
	}

	return genres, nil
}

func (pg *PostgreSQL) InsertGenres(ctx context.Context, genres types.Genres) error {
	genreBuilder := sq.Insert("genres").Columns("id", "name")
	for _, genre := range genres {
		genreBuilder = genreBuilder.Values(genre.ID, genre.Name)
	}
	sql1, args1, err := genreBuilder.PlaceholderFormat(sq.Dollar).Suffix("ON CONFLICT DO NOTHING").ToSql()
	if err != nil {
		return err
	}

	valueStrings := make([]string, 0, len(genres))
	for _, genre := range genres {
		valueStrings = append(valueStrings, fmt.Sprintf("('%s'::uuid, %d)", genre.ID, genre.TMDbID))
	}

	valuesSelect := sq.Select("t1.genre_id", "t1.tmdb_genre_id").
		From(fmt.Sprintf("(VALUES %s) AS t1(genre_id, tmdb_genre_id)", strings.Join(valueStrings, ", "))).
		Where("NOT EXISTS (SELECT 1 FROM link_tmdb_genres t2 WHERE t2.tmdb_genre_id = t1.tmdb_genre_id)")

	sb := sq.Insert("link_tmdb_genres").
		Columns("genre_id", "tmdb_genre_id").
		Select(valuesSelect).
		Suffix("ON CONFLICT DO NOTHING")

	sql2, args2, err := sb.ToSql()
	if err != nil {
		return err
	}

	tx, err := pg.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = pg.pool.Exec(ctx, sql1, args1...)
	if err != nil {
		return fmt.Errorf("failed to insert genre: %s", err.Error())
	}

	_, err = pg.pool.Exec(ctx, sql2, args2...)
	if err != nil {
		return fmt.Errorf("failed to insert link tmdb genre: %s", err.Error())
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (pg *PostgreSQL) InsertContentGenres(ctx context.Context, contentID uuid.UUID, tmdbGenreIDs []int64) error {
	if len(tmdbGenreIDs) == 0 {
		return nil
	}

	tx, err := pg.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	genreUUIDs, err := pg.getGenreUUIDsByTMDBIDs(ctx, tmdbGenreIDs)
	if err != nil {
		return err
	}
	builder := sq.Insert("link_content_genres").Columns("content_id", "genre_id")
	for _, genreID := range genreUUIDs {
		builder = builder.Values(contentID, genreID)
	}
	sql, args, err := builder.PlaceholderFormat(sq.Dollar).Suffix("ON CONFLICT DO NOTHING").ToSql()
	if err != nil {
		return err
	}
	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert link between film content and genre: %s", err.Error())
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (pg *PostgreSQL) getGenreUUIDsByTMDBIDs(ctx context.Context, tmdbGenreIDs []int64) ([]uuid.UUID, error) {
	sql, args, err := sq.Select("genre_id").
		From("link_tmdb_genres").
		Where("tmdb_genre_id = any(?)", tmdbGenreIDs).
		PlaceholderFormat(sq.Dollar).ToSql()
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
