package loader

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"whattowatch/internal/config"
	"whattowatch/internal/storage"
)

type ILoader interface {
	Load(ctx context.Context) error
}

type Loader struct {
	ID      int
	BaseUrl *url.URL
	log     *slog.Logger
}

func New(name string, cfg *config.Config, log *slog.Logger, storage storage.Storer) (ILoader, error) {
	source, err := storage.GetSourceByName(context.Background(), name)
	if err != nil {
		return nil, err
	}
	switch name {
	case "TMDb":
		return NewTMDbLoader(cfg.Tokens.TMDb, source.Url, log, storage)
	case "Kinopoisk":
		return NewKinopoiskLoader(cfg.Tokens.Kinopoisk, source.Url, log, storage)
	default:
		return nil, fmt.Errorf("unknown Loader name: %s", name)
	}
}
