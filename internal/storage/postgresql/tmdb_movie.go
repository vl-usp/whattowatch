package postgresql

import (
	"context"
	"fmt"
	"whattowatch/internal/types"

	sq "github.com/Masterminds/squirrel"
)

func (pg *PostgreSQL) GetTMDbMovieIDs(ctx context.Context) ([]int, error) {
	sql, args, err := sq.Select("id").PlaceholderFormat(sq.Dollar).From("tmdb.movies").ToSql()
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0)
	r, err := pg.pool.Query(ctx, sql, args...)
	if err != nil {
		return ids, fmt.Errorf("failed to get ids: %s", err.Error())
	}
	for r.Next() {
		var id int
		err = r.Scan(&id)
		if err != nil {
			return ids, fmt.Errorf("failed to scan: %s", err.Error())
		}
		ids = append(ids, id)
	}
	r.Close()
	return ids, nil
}

func (pg *PostgreSQL) UpdateTMDbMovie(ctx context.Context, movie types.TMDbMovie) error {
	sql, args, err := sq.Update("tmdb.movies").SetMap(sq.Eq{
		"overview":     movie.Overview,
		"popularity":   movie.Popularity,
		"poster_path":  movie.PosterPath,
		"release_date": movie.ReleaseDate,
		"budget":       movie.Budget,
		"revenue":      movie.Revenue,
		"runtime":      movie.Runtime,
		"vote_average": movie.VoteAverage,
		"vote_count":   movie.VoteCount,
	}).Where(sq.Eq{"id": movie.ID}).PlaceholderFormat(sq.Dollar).ToSql()

	if err != nil {
		return fmt.Errorf("failed to build update query: %s", err.Error())
	}
	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to update tmdb movie: %s", err.Error())
	}
	return nil
}

func (pg *PostgreSQL) InsertTMDbMovie(ctx context.Context, movie types.TMDbMovie) error {
	sql, args, err := sq.Insert("tmdb.movies").SetMap(sq.Eq{
		"id":           movie.ID,
		"title":        movie.Title,
		"overview":     movie.Overview,
		"popularity":   movie.Popularity,
		"poster_path":  movie.PosterPath,
		"release_date": movie.ReleaseDate,
		"budget":       movie.Budget,
		"revenue":      movie.Revenue,
		"runtime":      movie.Runtime,
		"vote_average": movie.VoteAverage,
		"vote_count":   movie.VoteCount,
	}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %s", err.Error())
	}

	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert tmdb movie: %s", err.Error())
	}
	return nil
}

func (pg *PostgreSQL) InsertTMDbMovies(ctx context.Context, movies []types.TMDbMovie) error {
	builder := sq.Insert("tmdb.movies").Columns(
		"id",
		"title",
		"overview",
		"popularity",
		"poster_path",
		"release_date",
		"budget",
		"revenue",
		"runtime",
		"vote_average",
		"vote_count",
	).PlaceholderFormat(sq.Dollar)

	for _, movie := range movies {
		if movie.ReleaseDate == "" {
			movie.ReleaseDate = "0001-01-01"
		}

		builder = builder.Values(
			movie.ID,
			movie.Title,
			movie.Overview,
			movie.Popularity,
			movie.PosterPath,
			movie.ReleaseDate,
			movie.Budget,
			movie.Revenue,
			movie.Runtime,
			movie.VoteAverage,
			movie.VoteCount,
		)
	}

	sql, args, err := builder.Suffix("ON CONFLICT DO NOTHING").ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %s", err.Error())
	}

	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert tmdb movie: %s", err.Error())
	}
	return nil
}
