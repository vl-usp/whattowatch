package storage

import (
	"context"
	"fmt"
	"log/slog"
	"whattowatch/internal/config"
	"whattowatch/internal/storage/postgresql"
	"whattowatch/internal/types"
)

type TMDbTVStorer interface {
	GetTMDbTVIDs(ctx context.Context) ([]int, error)
	UpdateTMDbTV(ctx context.Context, tv types.TMDbTV) error
	InsertTMDbTV(ctx context.Context, tv types.TMDbTV) error
	InsertTMDbTVs(ctx context.Context, tvs []types.TMDbTV) error
}

type TMDbMovieStorer interface {
	GetTMDbMovieIDs(ctx context.Context) ([]int, error)
	UpdateTMDbMovie(ctx context.Context, movie types.TMDbMovie) error
	InsertTMDbMovie(ctx context.Context, movie types.TMDbMovie) error
	InsertTMDbMovies(ctx context.Context, movies []types.TMDbMovie) error
}

type TMDbGenreStorer interface {
	InsertTMDbGenre(ctx context.Context, genre types.TMDbGenre) error
	InsertTMDbMoviesGenres(ctx context.Context, movieID int, genreIDs []int32) error
	InsertTMDbTVsGenres(ctx context.Context, tvID int, genreIDs []int32) error
	InsertTMDbGenres(ctx context.Context, genres []types.TMDbGenre) error
}

type SourceStorer interface {
	GetSources(ctx context.Context) ([]types.Source, error)
	GetSourceByName(ctx context.Context, name string) (*types.Source, error)
}

type Storer interface {
	SourceStorer
	TMDbTVStorer
	TMDbMovieStorer
	TMDbGenreStorer
}

func New(cfg *config.Config, log *slog.Logger) (Storer, error) {
	switch cfg.StorageType {
	case "postgresql":
		return postgresql.New(cfg, log)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", cfg.StorageType)
	}
}
