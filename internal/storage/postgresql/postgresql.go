package postgresql

import (
	"context"
	"log/slog"
	"whattowatch/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgreSQL struct {
	conn *pgxpool.Pool
	log  *slog.Logger
}

func New(cfg *config.Config, logger *slog.Logger) (*PostgreSQL, error) {
	pgxConfig, err := pgxpool.ParseConfig(cfg.DB.PostgresDSN)
	if err != nil {
		return nil, err
	}

	conn, err := pgxpool.NewWithConfig(context.Background(), pgxConfig)
	if err != nil {
		return nil, err
	}

	return &PostgreSQL{
		log:  logger.With("pkg", "postgresql"),
		conn: conn,
	}, nil
}
