package loader

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"whattowatch/internal/config"
	"whattowatch/internal/storage"
	"whattowatch/internal/types"
	"whattowatch/internal/utils"

	tmdbLib "github.com/cyruzin/golang-tmdb"
	"github.com/gofrs/uuid"
)

type Storer interface {
	storage.ContentStorer
	storage.GenreStorer
}

type TMDbLoader struct {
	BaseUrl *url.URL
	log     *slog.Logger
	storer  Storer
	client  *tmdbLib.Client
	cfg     *config.Config
	options map[string]string
}

func NewTMDbLoader(cfg *config.Config, log *slog.Logger, storer storage.Storer) (*TMDbLoader, error) {
	u, err := url.Parse(cfg.Urls.TMDbApiUrl)
	if err != nil {
		return nil, err
	}

	c, err := tmdbLib.Init(cfg.Tokens.TMDb)
	if err != nil {
		return nil, err
	}
	c.SetClientAutoRetry()
	// c.SetAlternateBaseURL()

	loader := &TMDbLoader{
		log:     log,
		BaseUrl: u,
		storer:  storer,
		client:  c,
		cfg:     cfg,
		options: make(map[string]string),
	}
	loader.options["language"] = "ru-RU"

	loader.log.Info("loader initialized", "url", cfg.Urls.TMDbApiUrl, "opts", loader.options)

	return loader, nil
}

func (l *TMDbLoader) Load(ctx context.Context) error {
	err := l.loadGenres(ctx)
	if err != nil {
		return fmt.Errorf("failed to load genres: %s", err.Error())
	}

	err1Ch := make(chan error)
	err2Ch := make(chan error)

	go func(errCh chan error) {
		l.log.Info("start discover and save movies")
		err := l.discoverAndSave(ctx, types.Movie)
		err1Ch <- err
	}(err1Ch)

	go func(errCh chan error) {
		l.log.Info("start discover and save tvs")
		err := l.discoverAndSave(ctx, types.TV)
		err2Ch <- err
	}(err2Ch)

	err1 := <-err1Ch
	l.log.Info("finish discover and save movies")
	if err1 != nil {
		return fmt.Errorf("failed to discover and save movies: %s", err1.Error())
	}

	err2 := <-err2Ch
	l.log.Info("finish discover and save tvs")
	if err2 != nil {
		return fmt.Errorf("failed to discover and save tvs: %s", err2.Error())
	}

	return nil
}

func (l *TMDbLoader) loadGenres(ctx context.Context) error {
	movieGenres, err := l.client.GetGenreMovieList(l.options)
	if err != nil {
		return fmt.Errorf("failed to get movie genres: %s", err.Error())
	}

	genres := make(map[int64]types.Genre, len(movieGenres.Genres))

	for _, genre := range movieGenres.Genres {
		id, err := uuid.NewV4()
		if err != nil {
			l.log.Error("failed to generate movie genre uuid", "error", err.Error())
			return err
		}
		genres[genre.ID] = types.Genre{ID: id, TMDbID: genre.ID, Name: genre.Name}
	}

	tvGenres, err := l.client.GetGenreTVList(l.options)
	if err != nil {
		return fmt.Errorf("failed to get tv genres: %s", err.Error())
	}

	for _, genre := range tvGenres.Genres {
		id, err := uuid.NewV4()
		if err != nil {
			l.log.Error("failed to generate tv genre uuid", "error", err.Error())
			return err
		}
		genres[genre.ID] = types.Genre{ID: id, TMDbID: genre.ID, Name: genre.Name}
	}

	l.log.Debug("genres insertion", "movies", movieGenres.Genres, "tvs", tvGenres.Genres, "total", genres)

	err = l.storer.InsertGenres(ctx, utils.MapToSlice(genres))
	if err != nil {
		return fmt.Errorf("failed to save genres: %s", err.Error())
	}

	return nil
}

func (l *TMDbLoader) discoverAndSave(ctx context.Context, dt types.ContentType) error {
	fromPage := 1
	toPage := 500
	for page := fromPage; page <= toPage; page++ {
		l.options["page"] = fmt.Sprintf("%d", page)

		switch dt {
		case types.Movie:
			res, err := l.client.GetDiscoverMovie(l.options)
			if err != nil {
				l.log.Error("failed to discover movies", "error", err.Error())
				return err
			}
			l.log.Info("discover movies success", "page", page, "total_pages", res.TotalPages, "total_results", res.TotalResults)
			movies := make(types.ContentSlice, 0, len(res.Results))
			moviesGenresMap := make(map[uuid.UUID][]int64, len(res.Results))
			for _, movie := range res.Results {
				releaseDate, err := types.GetReleaseDate(movie.ReleaseDate)
				if err != nil {
					l.log.Error("failed to get release date", "error", err.Error())
				}

				movieUUID, err := uuid.NewV4()
				if err != nil {
					l.log.Error("failed to generate movie uuid", "error", err.Error())
					return err
				}
				movies = append(movies, types.Content{
					ID:          movieUUID,
					ContentType: dt,
					Title:       movie.Title,
					Overview:    movie.Overview,
					Popularity:  movie.Popularity,
					PosterPath:  l.cfg.Urls.TMDbImageUrl + movie.PosterPath,
					ReleaseDate: releaseDate,
					VoteAverage: movie.VoteAverage,
					VoteCount:   movie.VoteCount,
					TMDbID:      movie.ID,
				})

				moviesGenresMap[movieUUID] = movie.GenreIDs
			}

			err = l.storer.InsertContentSlice(ctx, movies)
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
			res, err := l.client.GetDiscoverTV(l.options)
			if err != nil {
				l.log.Error("failed to discover TVs", "error", err.Error())
				return err
			}
			l.log.Info("discover TVs success", "page", page, "total_pages", res.TotalPages, "total_results", res.TotalResults)
			tvs := make(types.ContentSlice, 0, len(res.Results))
			tvsGenresMap := make(map[uuid.UUID][]int64, len(res.Results))
			for _, tv := range res.Results {
				tvTitle := tv.Name
				if tv.Name == "" {
					tvTitle = tv.OriginalName
				}

				releaseDate, err := types.GetReleaseDate(tv.FirstAirDate)
				if err != nil {
					l.log.Error("failed to get release date", "error", err.Error())
				}

				tvUUID, err := uuid.NewV4()
				if err != nil {
					l.log.Error("failed to generate movie uuid", "error", err.Error())
					return err
				}
				tvs = append(tvs, types.Content{
					ID:          tvUUID,
					TMDbID:      tv.ID,
					ContentType: dt,
					Title:       tvTitle,
					Overview:    tv.Overview,
					Popularity:  tv.Popularity,
					ReleaseDate: releaseDate,
					PosterPath:  l.cfg.Urls.TMDbImageUrl + tv.PosterPath,
					VoteAverage: tv.VoteAverage,
					VoteCount:   tv.VoteCount,
				})
				tvsGenresMap[tvUUID] = tv.GenreIDs
			}

			err = l.storer.InsertContentSlice(ctx, tvs)
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
