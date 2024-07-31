package loader

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"whattowatch/internal/storage"
	"whattowatch/internal/types"

	"github.com/ryanbradynd05/go-tmdb"
)

type TMDbLoader struct {
	Loader
	api     *tmdb.TMDb
	options map[string]string
}

func NewTMDbLoader(apiKey string, baseUrl string, log *slog.Logger, storage storage.IStorage) (*TMDbLoader, error) {
	config := tmdb.Config{
		APIKey:   apiKey,
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
	loader.options["language"] = "ru-RU"

	return loader, nil
}

func (l *TMDbLoader) Load(ctx context.Context) error {
	fromPage := 1
	toPage := 500
	for page := fromPage; page <= toPage; page++ {
		l.options["page"] = fmt.Sprintf("%d", page)

		res, err := l.api.DiscoverMovie(l.options)
		if err != nil {
			l.log.Error("failed to discover movies", "error", err.Error())
			return err
		}
		l.log.Info("request success", "page", page, "total_pages", res.TotalPages, "total_results", res.TotalResults)

		movies := make([]types.TMDbMovie, 0, len(res.Results))
		for _, movie := range res.Results {
			if movie.Overview == "" {
				continue
			}
			movies = append(movies, types.TMDbMovie{
				ID:          movie.ID,
				Title:       movie.Title,
				Overview:    movie.Overview,
				Popularity:  movie.Popularity,
				PosterPath:  movie.PosterPath,
				ReleaseDate: movie.ReleaseDate,
				VoteAverage: movie.VoteAverage,
				VoteCount:   movie.VoteCount,
			})
		}
		err = l.storage.InsertTMDbMovies(ctx, movies)
		if err != nil {
			l.log.Error("failed to insert movies", "error", err.Error(), "page", page, "movies", movies)
			return err
		}
	}
	return nil
}

func (l *TMDbLoader) UpdateData(ctx context.Context) error {
	ids, err := l.storage.GetTMDbMovieIDs(ctx)
	if err != nil {
		return err
	}

	for _, id := range ids {
		movie, err := l.api.GetMovieInfo(id, l.options)
		if err != nil {
			return err
		}

		m := types.TMDbMovie{
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
		}

		// TODO transaction
		err = l.storage.UpdateTMDbMovie(ctx, m)
		if err != nil {
			return err
		}

		for _, genre := range movie.Genres {
			err = l.storage.InsertTMDbGenre(ctx, types.TMDbGenre{
				Name: genre.Name,
				ID:   genre.ID,
			})
			if err != nil {
				return err
			}

			err = l.storage.InsertTMDbMovieGenre(ctx, genre.ID, movie.ID)
			if err != nil {
				return err
			}
		}

		l.log.Debug("loaded movie", "id", id, "movie", movie)
	}

	return nil
}
