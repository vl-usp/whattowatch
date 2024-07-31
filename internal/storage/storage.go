package storage

import (
	"context"
	"fmt"
	"log/slog"
	"whattowatch/internal/config"
	"whattowatch/internal/storage/postgresql"
	"whattowatch/internal/types"
)

type IStorage interface {
	InsertSourceLinks(ctx context.Context, links types.SourceLinkMap) error
	GetSources(ctx context.Context) ([]types.Source, error)
	GetSourceByName(ctx context.Context, name string) (*types.Source, error)
}

func New(cfg *config.Config, log *slog.Logger) (IStorage, error) {
	switch cfg.StorageType {
	case "postgresql":
		return postgresql.New(cfg, log)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", cfg.StorageType)
	}
}
