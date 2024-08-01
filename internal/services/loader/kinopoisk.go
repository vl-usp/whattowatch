package loader

import (
	"context"
	"log/slog"
	"net/url"
	"whattowatch/internal/storage"
)

type KinopoiskLoader struct {
	Loader
	apiKey string
	limit  int
}

func NewKinopoiskLoader(apiKey string, baseUrl string, log *slog.Logger, storer storage.Storer) (*KinopoiskLoader, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	return &KinopoiskLoader{
		Loader: Loader{
			log:     log,
			storer:  storer,
			BaseUrl: u,
		},
		apiKey: apiKey,
		limit:  500,
	}, nil
}

func (p *KinopoiskLoader) Load(ctx context.Context) error {
	return nil
}
