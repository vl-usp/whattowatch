package loader

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"whattowatch/internal/storage"
	"whattowatch/internal/types"

	"github.com/gofrs/uuid"
	tmdbLib "github.com/ryanbradynd05/go-tmdb"
)

type Storer interface {
	storage.ContentStorer
	storage.GenreStorer
}

type TMDbLoader struct {
	BaseUrl *url.URL
	log     *slog.Logger
	storer  Storer
	api     *tmdbLib.TMDb
	options map[string]string
}

func NewTMDbLoader(apiKey string, baseUrl string, log *slog.Logger, storer storage.Storer) (*TMDbLoader, error) {
	config := tmdbLib.Config{
		APIKey: apiKey,
	}
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	loader := &TMDbLoader{
		log:     log,
		BaseUrl: u,
		storer:  storer,
		api:     tmdbLib.Init(config),
		options: make(map[string]string),
	}
	// loader.options["language"] = "ru-RU"

	loader.log.Info("loader initialized", "url", baseUrl, "opts", loader.options)

	return loader, nil
}

func (l *TMDbLoader) Load(ctx context.Context) error {
	movieGenres, err := l.api.GetMovieGenres(l.options)
	if err != nil {
		return fmt.Errorf("failed to get movie genres: %s", err.Error())
	}
	l.log.Debug("get movie genres", "genres", movieGenres.Genres)
	genres := make(types.Genres, 0, len(movieGenres.Genres))
	for _, genre := range movieGenres.Genres {
		id, err := uuid.NewV7()
		if err != nil {
			l.log.Error("failed to generate movie uuid", "error", err.Error())
			return err
		}
		genres = append(genres, types.Genre{ID: id, TMDbID: genre.ID, Name: genre.Name})
	}
	tvGenres, err := l.api.GetTvGenres(l.options)
	if err != nil {
		return fmt.Errorf("failed to save tv genres: %s", err.Error())
	}
	for _, genre := range tvGenres.Genres {
		id, err := uuid.NewV7()
		if err != nil {
			l.log.Error("failed to generate movie uuid", "error", err.Error())
			return err
		}
		genres = append(genres, types.Genre{ID: id, TMDbID: genre.ID, Name: genre.Name})
	}
	err = l.storer.InsertGenres(ctx, genres)
	if err != nil {
		return fmt.Errorf("failed to save tv genres: %s", err.Error())
	}

	err1Ch := make(chan error)
	err2Ch := make(chan error)
	go func(errCh chan error) {
		l.log.Info("start discover and save movies")
		err := l.DiscoverAndSave(ctx, types.Movie)
		err1Ch <- err
	}(err1Ch)
	go func(errCh chan error) {
		l.log.Info("start discover and save tvs")
		err := l.DiscoverAndSave(ctx, types.TV)
		err2Ch <- err
	}(err2Ch)

	err1 := <-err1Ch
	l.log.Info("finish discover and save movies")
	if err1 != nil {
		return fmt.Errorf("failed to discover movies: %s", err1.Error())
	}
	err2 := <-err2Ch
	l.log.Info("finish discover and save tvs")
	if err2 != nil {
		return fmt.Errorf("failed to discover tvs: %s", err2.Error())
	}
	return nil
}

func (l *TMDbLoader) DiscoverAndSave(ctx context.Context, dt types.ContentType) error {
	fromPage := 1
	toPage := 500
	for page := fromPage; page <= toPage; page++ {
		l.options["page"] = fmt.Sprintf("%d", page)

		switch dt {
		case types.Movie:
			res, err := l.api.DiscoverMovie(l.options)
			if err != nil {
				l.log.Error("failed to discover movies", "error", err.Error())
				return err
			}
			l.log.Info("discover movies success", "page", page, "total_pages", res.TotalPages, "total_results", res.TotalResults)
			movies := make(types.Contents, 0, len(res.Results))
			moviesGenresMap := make(map[uuid.UUID][]int32, len(res.Results))
			for _, movie := range res.Results {
				releaseDate, err := types.GetReleaseDate(movie.ReleaseDate)
				if err != nil {
					l.log.Error("failed to get release date", "error", err.Error())
				}
				// movieUUID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(movie.Title))
				movieUUID, err := uuid.NewV7()
				if err != nil {
					l.log.Error("failed to generate movie uuid", "error", err.Error())
					return err
				}
				movies = append(movies, types.Content{
					ID:            movieUUID,
					ContentTypeID: dt.EnumIndex(),
					Title:         movie.Title,
					Overview:      movie.Overview,
					Popularity:    movie.Popularity,
					PosterPath:    movie.PosterPath,
					ReleaseDate:   releaseDate,
					VoteAverage:   movie.VoteAverage,
					VoteCount:     movie.VoteCount,
					TMDbID:        movie.ID,
				})

				moviesGenresMap[movieUUID] = movie.GenreIDs
			}

			err = l.storer.InsertContents(ctx, movies)
			if err != nil {
				l.log.Error("failed to insert movies", "error", err.Error(), "page", page, "movies", movies)
				return err
			}

			for movieID, genreIDs := range moviesGenresMap {
				err = l.storer.InsertContentGenres(ctx, movieID, genreIDs)
				if err != nil {
					l.log.Error("failed to insert movies genres", "error", err.Error(), "movieID", movieID, "genreIDs", genreIDs)
					return err
				}
			}
		case types.TV:
			res, err := l.api.DiscoverTV(l.options)
			if err != nil {
				l.log.Error("failed to discover TVs", "error", err.Error())
				return err
			}
			l.log.Info("discover TVs success", "page", page, "total_pages", res.TotalPages, "total_results", res.TotalResults)
			tvs := make(types.Contents, 0, len(res.Results))
			tvsGenresMap := make(map[uuid.UUID][]int32, len(res.Results))
			for _, tv := range res.Results {
				tvTitle := tv.Name
				if tv.Name == "" {
					tvTitle = tv.OriginalName
				}

				releaseDate, err := types.GetReleaseDate(tv.FirstAirDate)
				if err != nil {
					l.log.Error("failed to get release date", "error", err.Error())
				}

				tvUUID, err := uuid.NewV7()
				if err != nil {
					l.log.Error("failed to generate movie uuid", "error", err.Error())
					return err
				}
				tvs = append(tvs, types.Content{
					ID:            tvUUID,
					TMDbID:        tv.ID,
					ContentTypeID: dt.EnumIndex(),
					Title:         tvTitle,
					Overview:      tv.Overview,
					Popularity:    tv.Popularity,
					ReleaseDate:   releaseDate,
					PosterPath:    tv.PosterPath,
					VoteAverage:   tv.VoteAverage,
					VoteCount:     tv.VoteCount,
				})
				tvsGenresMap[tvUUID] = tv.GenreIDs
			}

			err = l.storer.InsertContents(ctx, tvs)
			if err != nil {
				l.log.Error("failed to insert tvs", "error", err.Error(), "page", page, "tvs", tvs)
				return err
			}

			for tvID, genreIDs := range tvsGenresMap {
				err = l.storer.InsertContentGenres(ctx, tvID, genreIDs)
				if err != nil {
					l.log.Error("failed to insert tvs genres", "error", err.Error(), "tvID", tvID, "genreIDs", genreIDs)
					return err
				}
			}
		}
	}
	return nil
}
