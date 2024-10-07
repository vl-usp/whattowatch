package api

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
	"whattowatch/internal/storage"
	"whattowatch/internal/types"

	"github.com/gofrs/uuid"
	"github.com/ryanbradynd05/go-tmdb"
)

type TMDbApi struct {
	api    *tmdb.TMDb
	opts   map[string]string
	storer storage.Storer

	log *slog.Logger
}

func New(apiKey string, storer storage.Storer, log *slog.Logger) *TMDbApi {
	opts := make(map[string]string)
	opts["language"] = "ru-RU"
	return &TMDbApi{
		api:    tmdb.Init(tmdb.Config{APIKey: apiKey, Proxies: nil, UseProxy: false}),
		opts:   opts,
		storer: storer,

		log: log.With("pkg", "api"),
	}
}

func (api *TMDbApi) GetMoviesRecomendations(ctx context.Context, movies types.Contents) (types.Contents, error) {
	log := api.log.With("fn", "getMoviesRecomendation")
	contents := make(types.Contents, 0)
	for _, m := range movies {
		recs, err := api.api.GetMovieRecommendations(m.TMDbID, api.opts)
		if err != nil {
			log.Error("failed to get movie recomendations", "err", err.Error())
			return contents, err
		}

		for _, rec := range recs.Results {
			genres, err := api.storer.GetGenresByIDs(ctx, rec.GenreIDs)
			if err != nil {
				log.Error("failed to get movie genres", "err", err.Error())
				return contents, err
			}
			releaseDate, err := time.Parse("2006-01-02", rec.ReleaseDate)
			if err != nil {
				log.Error("failed to parse release date", "err", err.Error())
				return contents, err
			}
			contents = append(contents, types.Content{
				ID:          uuid.Nil,
				TMDbID:      rec.ID,
				Title:       rec.Title,
				Genres:      genres,
				Overview:    rec.Overview,
				ReleaseDate: sql.NullTime{Time: releaseDate, Valid: true},
				PosterPath:  rec.PosterPath,
			})
		}
	}
	return contents, nil
}

func (api *TMDbApi) GetTVsRecomendations(ctx context.Context, tvs types.Contents) (types.Contents, error) {
	log := api.log.With("fn", "getTVsRecomendation")
	contents := make(types.Contents, 0)
	for _, m := range tvs {
		recs, err := api.api.GetTvRecommendations(m.TMDbID, api.opts)
		if err != nil {
			log.Error("failed to get tvs recomendations", "err", err.Error())
			return contents, err
		}

		for _, rec := range recs.Results {
			genres, err := api.storer.GetGenresByIDs(ctx, rec.GenreIDs)
			if err != nil {
				log.Error("failed to get tvs genres", "err", err.Error())
				return contents, err
			}
			releaseDate, err := time.Parse("2006-01-02", rec.FirstAirDate)
			if err != nil {
				log.Error("failed to parse release date", "err", err.Error())
				return contents, err
			}
			contents = append(contents, types.Content{
				ID:          uuid.Nil,
				TMDbID:      rec.ID,
				Title:       rec.Name,
				Genres:      genres,
				Overview:    rec.Overview,
				ReleaseDate: sql.NullTime{Time: releaseDate, Valid: true},
				PosterPath:  rec.PosterPath,
			})
		}
	}
	return contents, nil
}

func (api *TMDbApi) GetTopMovies(ctx context.Context) (types.Contents, error) {
	return nil, nil
}
