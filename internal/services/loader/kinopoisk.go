package loader

import (
	"context"
	"log/slog"
	"net/url"
	"whattowatch/internal/storage"
)

type KinopoiskLoader struct {
	Loader
	storer storage.KinopoiskStorer
	apiKey string
	limit  int
}

func NewKinopoiskLoader(apiKey string, baseUrl string, log *slog.Logger, storer storage.KinopoiskStorer) (*KinopoiskLoader, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	return &KinopoiskLoader{
		Loader: Loader{
			log:     log,
			BaseUrl: u,
		},
		storer: storer,
		apiKey: apiKey,
		limit:  500,
	}, nil
}

func (p *KinopoiskLoader) Load(ctx context.Context) error {
	// TODO: impliment me
	return nil
}
