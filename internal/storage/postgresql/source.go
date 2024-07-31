package postgresql

import (
	"context"
	"fmt"
	"whattowatch/internal/types"

	sq "github.com/Masterminds/squirrel"
)

func (pg *PostgreSQL) GetSources(ctx context.Context) ([]types.Source, error) {
	sql, args, err := sq.Select("*").PlaceholderFormat(sq.Dollar).From("sources").ToSql()
	if err != nil {
		return nil, err
	}
	var sources []types.Source
	rows, err := pg.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var source types.Source
		err = rows.Scan(&source.ID, &source.Name, &source.Url)
		if err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}
	return sources, nil
}

func (pg *PostgreSQL) GetSourceByName(ctx context.Context, name string) (*types.Source, error) {
	sql, args, err := sq.Select("id", "url").PlaceholderFormat(sq.Dollar).From("sources").Where(sq.Eq{"name": name}).ToSql()
	if err != nil {
		return nil, err
	}
	var id *int
	var url *string
	err = pg.pool.QueryRow(ctx, sql, args...).Scan(&id, &url)
	if err != nil {
		return nil, fmt.Errorf("failed to get source: %s", err.Error())
	}
	return &types.Source{ID: *id, Name: name, Url: *url}, nil
}
