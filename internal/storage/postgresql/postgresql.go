package postgresql

import (
	"context"
	"log/slog"
	"whattowatch/internal/config"

	pgxUUID "github.com/jackc/pgx-gofrs-uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgreSQL struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

func New(cfg *config.Config, log *slog.Logger) (*PostgreSQL, error) {
	pgxConfig, err := pgxpool.ParseConfig(cfg.DB.PostgresDSN)
	if err != nil {
		return nil, err
	}

	pgxConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), pgxConfig)
	if err != nil {
		return nil, err
	}

	return &PostgreSQL{
		log:  log,
		pool: pool,
	}, nil
}
