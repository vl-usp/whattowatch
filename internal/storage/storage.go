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
	GetSources(ctx context.Context) ([]types.Source, error)
	GetSourceByName(ctx context.Context, name string) (*types.Source, error)
	InsertTMDbMovies(ctx context.Context, movies []types.TMDbMovie) error
	InsertTMDbGenre(ctx context.Context, genre types.TMDbGenre) error
	InsertTMDbMovieGenre(ctx context.Context, genreID, movieID int) error
	GetTMDbMovieIDs(ctx context.Context) ([]int, error)
	UpdateTMDbMovie(ctx context.Context, movie types.TMDbMovie) error
}

func New(cfg *config.Config, log *slog.Logger) (IStorage, error) {
	switch cfg.StorageType {
	case "postgresql":
		return postgresql.New(cfg, log)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", cfg.StorageType)
	}
}
