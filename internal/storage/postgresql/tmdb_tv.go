package postgresql

import (
	"context"
	"fmt"
	"whattowatch/internal/types"

	sq "github.com/Masterminds/squirrel"
)

func (pg *PostgreSQL) GetTMDbTVIDs(ctx context.Context) ([]int, error) {
	sql, args, err := sq.Select("id").PlaceholderFormat(sq.Dollar).From("tmdb.tvs").ToSql()
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

func (pg *PostgreSQL) UpdateTMDbTV(ctx context.Context, tv types.TMDbTV) error {
	sql, args, err := sq.Update("tmdb.movies").SetMap(sq.Eq{
		"overview":       tv.Overview,
		"popularity":     tv.Popularity,
		"poster_path":    tv.PosterPath,
		"first_air_date": tv.FirstAirDate,
		"vote_average":   tv.VoteAverage,
		"vote_count":     tv.VoteCount,
	}).Where(sq.Eq{"id": tv.ID}).PlaceholderFormat(sq.Dollar).ToSql()

	if err != nil {
		return fmt.Errorf("failed to build update query: %s", err.Error())
	}
	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to update tmdb tv: %s", err.Error())
	}
	return nil
}

func (pg *PostgreSQL) InsertTMDbTV(ctx context.Context, tv types.TMDbTV) error {
	sql, args, err := sq.Insert("tmdb.movies").SetMap(sq.Eq{
		"id":             tv.ID,
		"title":          tv.Title,
		"overview":       tv.Overview,
		"popularity":     tv.Popularity,
		"poster_path":    tv.PosterPath,
		"first_air_date": tv.FirstAirDate,
		"vote_average":   tv.VoteAverage,
		"vote_count":     tv.VoteCount,
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

func (pg *PostgreSQL) InsertTMDbTVs(ctx context.Context, tvs []types.TMDbTV) error {
	builder := sq.Insert("tmdb.tvs").Columns(
		"id",
		"title",
		"overview",
		"popularity",
		"poster_path",
		"first_air_date",
		"vote_average",
		"vote_count",
	).PlaceholderFormat(sq.Dollar)

	for _, tv := range tvs {
		if tv.FirstAirDate == "" {
			tv.FirstAirDate = "0001-01-01"
		}

		builder = builder.Values(
			tv.ID,
			tv.Title,
			tv.Overview,
			tv.Popularity,
			tv.PosterPath,
			tv.FirstAirDate,
			tv.VoteAverage,
			tv.VoteCount,
		)
	}

	sql, args, err := builder.Suffix("ON CONFLICT DO NOTHING").ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %s", err.Error())
	}

	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert tmdb tv: %s", err.Error())
	}
	return nil
}
