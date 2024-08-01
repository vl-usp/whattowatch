package recomendation

import (
	"context"
	"log/slog"
	tmdb_api "whattowatch/internal/api/tmdb"
	"whattowatch/internal/config"
	"whattowatch/internal/storage"
)

type Service struct {
	storer storage.TMDbStorer
	api    *tmdb_api.TMDbApi

	cfg *config.Config
	log *slog.Logger
}

func New(cfg *config.Config, log *slog.Logger, storer storage.TMDbStorer) (*Service, error) {
	api, err := tmdb_api.New(cfg.Tokens.TMDb, log)
	if err != nil {
		return nil, err
	}
	return &Service{
		storer: storer,
		api:    api,
		cfg:    cfg,
		log:    log,
	}, nil
}

func (t *Service) GetRecomendationsFromTMDb(ctx context.Context, name string) error {
	// TODO: impliment me
	// get movie info by name
	// get tv info by name
	// get genre info
	// get recommendations
	return nil
}
