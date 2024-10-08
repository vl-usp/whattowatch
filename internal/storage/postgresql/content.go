package postgresql

import (
	"context"
	"fmt"
	"whattowatch/internal/types"

	sq "github.com/Masterminds/squirrel"
)

func (pg *PostgreSQL) GetContent(ctx context.Context, id int64) (types.Content, error) {
	sql, args, err := sq.Select("*").PlaceholderFormat(sq.Dollar).From("content").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return types.Content{}, err
	}
	var fc types.Content
	err = pg.pool.QueryRow(ctx, sql, args...).Scan(
		&fc.ID,
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

	sql, args, err := builder.Suffix("ON CONFLICT DO NOTHING").ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert content query: %s", err.Error())
	}

	_, err = pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert content: %s", err.Error())
	}

	return nil
}
