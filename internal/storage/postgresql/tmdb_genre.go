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

func (pg *PostgreSQL) InsertTMDbMovieGenre(ctx context.Context, genreID, movieID int) error {
	sql, args, err := sq.Insert("tmdb.movies_genres").Columns("movie_id", "genre_id").Values(movieID, genreID).PlaceholderFormat(sq.Dollar).Suffix("ON CONFLICT DO NOTHING").ToSql()
	if err != nil {
		return err
	}
	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert link between movie and genre: %s", err.Error())
	}
	return nil
}
