package postgresql

import (
	"context"
	"log/slog"
	"time"
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, err
	}

	err = conn.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return &PostgreSQL{
		log:  logger.With("pkg", "postgresql"),
		conn: conn,
	}, nil
}
