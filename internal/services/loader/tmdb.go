package loader

import (
	"log/slog"
	"net/url"
	"whattowatch/internal/storage"

	"github.com/ryanbradynd05/go-tmdb"
)

type TMDbLoader struct {
	Loader
	api     *tmdb.TMDb
	options map[string]string
}

func NewTMDbLoader(apiKey string, baseUrl string, log *slog.Logger, storage storage.IStorage) (*TMDbLoader, error) {
	config := tmdb.Config{
		APIKey:   "YOUR_KEY",
		Proxies:  nil,
		UseProxy: false,
	}
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	loader := &TMDbLoader{
		Loader: Loader{
			log:     log,
			storage: storage,
			BaseUrl: u,
		},
		api:     tmdb.Init(config),
		options: make(map[string]string),
	}
	loader.options["language"] = "RU"

	return loader, nil
}

func (l *TMDbLoader) Load() error {
	return nil
}
