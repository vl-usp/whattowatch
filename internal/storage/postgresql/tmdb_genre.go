package postgresql

import (
	"context"
	"fmt"
	"whattowatch/internal/types"

	sq "github.com/Masterminds/squirrel"
)

func (pg *PostgreSQL) InsertTMDbGenre(ctx context.Context, genre types.TMDbGenre) error {
	sql, args, err := sq.Insert("tmdb.genres").Columns("id", "name").Values(genre.ID, genre.Name).PlaceholderFormat(sq.Dollar).Suffix("ON CONFLICT DO NOTHING").ToSql()
	if err != nil {
		return err
	}
	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert tmdb genre: %s", err.Error())
	}
	return nil
}

func (pg *PostgreSQL) InsertTMDbMoviesGenres(ctx context.Context, movieID int, genreIDs []int32) error {
	if len(genreIDs) == 0 {
		return nil
	}
	builder := sq.Insert("tmdb.movies_genres").Columns("movie_id", "genre_id")
	for _, genreID := range genreIDs {
		builder = builder.Values(movieID, genreID)
	}
	sql, args, err := builder.PlaceholderFormat(sq.Dollar).Suffix("ON CONFLICT DO NOTHING").ToSql()
	if err != nil {
		return err
	}
	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert link between movie and genre: %s", err.Error())
	}
	return nil
}

func (pg *PostgreSQL) InsertTMDbTVsGenres(ctx context.Context, tvID int, genreIDs []int32) error {
	if len(genreIDs) == 0 {
		return nil
	}
	builder := sq.Insert("tmdb.tvs_genres").Columns("tv_id", "genre_id")
	for _, genreID := range genreIDs {
		builder = builder.Values(tvID, genreID)
	}
	sql, args, err := builder.PlaceholderFormat(sq.Dollar).Suffix("ON CONFLICT DO NOTHING").ToSql()
	if err != nil {
		return err
	}
	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert link between tvs and genre: %s", err.Error())
	}
	return nil
}

func (pg *PostgreSQL) InsertTMDbGenres(ctx context.Context, genres []types.TMDbGenre) error {
	builder := sq.Insert("tmdb.genres").Columns("id", "name")
	for _, genre := range genres {
		builder = builder.Values(genre.ID, genre.Name)
	}
	sql, args, err := builder.PlaceholderFormat(sq.Dollar).Suffix("ON CONFLICT DO NOTHING").ToSql()
	if err != nil {
		return err
	}
	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert tmdb genre: %s", err.Error())
	}
	return nil
}
