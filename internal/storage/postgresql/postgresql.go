package postgresql

import (
	"context"
	"log/slog"
	"whattowatch/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgreSQL struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

func New(cfg *config.Config, log *slog.Logger) (*PostgreSQL, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DB.PostgresDSN)
	if err != nil {
		return nil, err
	}

	return &PostgreSQL{
		log:  log,
		pool: pool,
	}, nil
}
