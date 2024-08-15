package api

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
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

func (api *TMDbApi) GetAllRecomendations(ctx context.Context, userID int) (types.ContentsByTypes, error) {
	favorites, err := api.storer.GetUserFavoritesByType(ctx, userID)
	if err != nil {
		return nil, err
	}

	tvsCh := make(chan types.Contents, 20)
	moviesCh := make(chan types.Contents, 20)
	errCh := make(chan error)
	for contentType, contents := range favorites {
		switch contentType {
		case types.MovieContentType:
			go func() {
				movies, err := api.GetMoviesRecomendations(ctx, contents)
				if err != nil {
					errCh <- err
				}
				moviesCh <- movies
			}()
		case types.TVContentType:
			go func() {
				tvs, err := api.GetTVsRecomendations(ctx, contents)
				if err != nil {
					errCh <- err
				}
				tvsCh <- tvs
			}()
		}
	}

	m := make(types.ContentsByTypes)
	mu := &sync.Mutex{}
	for i := 0; i < 2; i++ {
		go func() {
			for {
				select {
				case err := <-errCh:
					api.log.Error("failed to get recomendations", "err", err)
				case content := <-moviesCh:
					mu.Lock()
					m[types.MovieContentType] = append(m[types.MovieContentType], content...)
					mu.Unlock()
				case content := <-tvsCh:
					mu.Lock()
					m[types.TVContentType] = append(m[types.TVContentType], content...)
					mu.Unlock()
				}
			}
		}()
	}

	return m, nil
}

func (api *TMDbApi) GetRecomendations(ctx context.Context, userID int, contentType types.ContentType) (types.ContentsByGenres, error) {
	log := api.log.With("fn", "GetRecomendations", "userID", userID, "contentType", contentType)
	favorites, err := api.storer.GetUserFavoritesByType(ctx, userID)
	if err != nil {
		log.Error("failed to get recomendations", "error", err.Error())
		return nil, err
	}

	if len(favorites[contentType]) == 0 {
		log.Error("failed to get recomendations", "error", "user favorites not found. please, add favorites before use this command")
		return nil, fmt.Errorf("user favorites not found. please, add favorites before use this command")
	}

	var recs types.Contents
	switch contentType {
	case types.MovieContentType:
		recs, err = api.GetMoviesRecomendations(ctx, favorites[contentType])
		if err != nil {
			log.Error("failed to get movies recomendations", "error", err.Error())
			return nil, err
		}
	case types.TVContentType:
		recs, err = api.GetTVsRecomendations(ctx, favorites[contentType])
		if err != nil {
			log.Error("failed to get tvs recomendations", "error", err.Error())
			return nil, err
		}
	default:
		log.Error("failed to get recomendations", "error", "content type not found")
		return nil, fmt.Errorf("content type not found")
	}

	result := make(types.ContentsByGenres, 0)
	for _, rec := range recs {
		if len(rec.Genres) > 0 {
			result[rec.Genres[0].Name] = append(result[rec.Genres[0].Name], rec)
		}
	}

	return result, nil
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
