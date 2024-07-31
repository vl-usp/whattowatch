package postgresql

import (
	"context"
	"whattowatch/internal/types"

	sq "github.com/Masterminds/squirrel"
)

func (pg *PostgreSQL) InsertSourceLinks(ctx context.Context, links types.SourceLinkMap) error {
	if len(links) == 0 {
		return nil
	}
	builderInsert := sq.Insert("source_links").PlaceholderFormat(sq.Dollar).Columns("source_id", "original_id", "title", "page", "movie_url")

	for _, link := range links {
		builderInsert = builderInsert.Values(link.SourceID, link.OriginalID, link.Title, link.Page, link.MovieUrl)
	}

	sql, args, err := builderInsert.Suffix("ON CONFLICT DO NOTHING").ToSql()
	if err != nil {
		return err
	}
	r, err := pg.pool.Query(ctx, sql, args...)
	if err != nil {
		return err
	}
	r.Close()
	return nil
}
