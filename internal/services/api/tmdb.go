package api

import (
	"log/slog"
	"whattowatch/internal/types"

	"github.com/ryanbradynd05/go-tmdb"
)

type TMDbApi struct {
	log     *slog.Logger
	api     *tmdb.TMDb
	options map[string]string
}

func NewTMDbApi(apiKey string, log *slog.Logger) (*TMDbApi, error) {
	config := tmdb.Config{
		APIKey:   apiKey,
		Proxies:  nil,
		UseProxy: false,
	}

	loader := &TMDbApi{
		log:     log,
		api:     tmdb.Init(config),
		options: make(map[string]string),
	}
	loader.options["language"] = "ru-RU"

	return loader, nil
}

func (l *TMDbApi) GetMovieInfo(id int) (*types.TMDbMovie, error) {
	movie, err := l.api.GetMovieInfo(id, l.options)
	if err != nil {
		return nil, err
	}
	return &types.TMDbMovie{
		ID:          movie.ID,
		Title:       movie.Title,
		Overview:    movie.Overview,
		Popularity:  movie.Popularity,
		PosterPath:  movie.PosterPath,
		ReleaseDate: movie.ReleaseDate,
		Budget:      movie.Budget,
		Revenue:     movie.Revenue,
		Runtime:     movie.Runtime,
		VoteAverage: movie.VoteAverage,
		VoteCount:   movie.VoteCount,
	}, nil
}
