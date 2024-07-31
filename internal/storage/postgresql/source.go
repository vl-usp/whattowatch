package postgresql

import (
	"context"
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
		err = rows.Scan(&source.ID, &source.Name, &source.Url, &source.CreatedAt, &source.UpdatedAt, &source.DeletedAt)
		if err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}
	return sources, nil
}

func (pg *PostgreSQL) GetSourceByName(ctx context.Context, name string) (*types.Source, error) {
	sql, args, err := sq.Select("*").PlaceholderFormat(sq.Dollar).From("sources").Where(sq.Eq{"name": name}).ToSql()
	if err != nil {
		return nil, err
	}
	var source *types.Source
	err = pg.pool.QueryRow(ctx, sql, args...).Scan(&source)
	if err != nil {
		return nil, err
	}
	return source, nil
}
