package postgresql

import (
	"context"
	"fmt"
	"strings"
	"whattowatch/internal/types"

	sq "github.com/Masterminds/squirrel"
)

func (pg *PostgreSQL) GetGenres(ctx context.Context, contentID int64) (types.Genres, error) {
	sql, args, err := sq.Select("t1.id", "t1.name", "t1.pretty_name").
		From("genres t1").
		Join("link_content_genres t2 on t1.id = t2.genre_id").
		Where("t2.content_id = ?", contentID).PlaceholderFormat(sq.Dollar).ToSql()

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
		err = rows.Scan(&genre.ID, &genre.Name, &genre.PrettyName)
		if err != nil {
			return nil, err
		}
		genres = append(genres, genre)
	}

	return genres, nil
}

func (pg *PostgreSQL) GetGenresByIDs(ctx context.Context, ids []int) (types.Genres, error) {
	sql, args, err := sq.Select("t1.id", "t1.name", "t1.pretty_name").
		From("genres t1").
		Where("t1.id = any(?)", ids).PlaceholderFormat(sq.Dollar).ToSql()
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
		err = rows.Scan(&genre.ID, &genre.Name, &genre.PrettyName)
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
	sql, args, err := genreBuilder.PlaceholderFormat(sq.Dollar).Suffix("ON CONFLICT DO NOTHING").ToSql()
	if err != nil {
		return err
	}

	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert genre: %s", err.Error())
	}

	return nil
}

func (pg *PostgreSQL) InsertContentGenres(ctx context.Context, contentID int64, genreIDs []int64) error {
	if len(genreIDs) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(genreIDs))
	for _, genreID := range genreIDs {
		valueStrings = append(valueStrings, fmt.Sprintf("(%d, %d)", contentID, genreID))
	}

	valuesSelect := sq.Select("t1.content_id", "t1.genre_id").
		From(fmt.Sprintf("(VALUES %s) AS t1(content_id, genre_id)", strings.Join(valueStrings, ", "))).
		Where("EXISTS (SELECT 1 FROM genres t2 WHERE t2.id = t1.genre_id)").
		Where("EXISTS (SELECT 1 FROM content t3 WHERE t3.id = t1.content_id)")

	sb := sq.Insert("link_content_genres").
		Columns("content_id", "genre_id").
		Select(valuesSelect).
		Suffix("ON CONFLICT DO NOTHING")

	sql, args, err := sb.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return err
	}

	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert link between film content and genre: %s", err.Error())
	}

	return nil
}
