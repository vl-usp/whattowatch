package postgresql

import (
	"context"
	"fmt"
	"strings"
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
		&fc.ContentType,
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

func (pg *PostgreSQL) InsertContentSlice(ctx context.Context, contents types.ContentSlice) error {
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
			c.ContentType.EnumIndex(),
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
		return fmt.Errorf("failed to build insert content query: %s", err.Error())
	}

	valueStrings := make([]string, 0, len(contents))
	for _, c := range contents {
		valueStrings = append(valueStrings, fmt.Sprintf("('%s'::uuid, %d)", c.ID, c.TMDbID))
	}

	valuesSelect := sq.Select("t1.content_id", "t1.tmdb_content_id").
		From(fmt.Sprintf("(VALUES %s) AS t1(content_id, tmdb_content_id)", strings.Join(valueStrings, ", "))).
		Where("NOT EXISTS (SELECT 1 FROM link_tmdb_content t2 WHERE t2.tmdb_content_id = t1.tmdb_content_id)")

	sb := sq.Insert("link_tmdb_content").
		Columns("content_id", "tmdb_content_id").
		Select(valuesSelect).
		Suffix("ON CONFLICT DO NOTHING")

	sql2, args2, err := sb.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert link_tmdb_content query: %s", err.Error())
	}

	tx, err := pg.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = pg.pool.Exec(ctx, sql1, args1...)
	if err != nil {
		return fmt.Errorf("failed to insert content: %s", err.Error())
	}

	_, err = pg.pool.Exec(ctx, sql2, args2...)
	if err != nil {
		return fmt.Errorf("failed to insert tmdb link content: %s", err.Error())
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
