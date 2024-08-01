package loader

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"whattowatch/internal/storage"
	"whattowatch/internal/types"

	tmdbLib "github.com/ryanbradynd05/go-tmdb"
)

type DiscoverType int

const (
	Movie DiscoverType = iota + 1
	TV
)

func (w DiscoverType) String() string {
	return [...]string{"Movie", "TV"}[w-1]
}

func (w DiscoverType) EnumIndex() int {
	return int(w)
}

type TMDbLoader struct {
	Loader
	storer  storage.TMDbStorer
	api     *tmdbLib.TMDb
	options map[string]string
}

func NewTMDbLoader(apiKey string, baseUrl string, log *slog.Logger, storer storage.Storer) (*TMDbLoader, error) {
	config := tmdbLib.Config{
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
			BaseUrl: u,
		},
		storer:  storer,
		api:     tmdbLib.Init(config),
		options: make(map[string]string),
	}
	loader.options["language"] = "ru-RU"
	loader.options["include_adult"] = "true"
	loader.options["sort_by"] = "vote_average.desc"

	return loader, nil
}

func (l *TMDbLoader) Load(ctx context.Context) error {
	// load genres before discover
	movieGenres, err := l.api.GetMovieGenres(l.options)
	if err != nil {
		return fmt.Errorf("failed to get movie genres: %s", err.Error())
	}
	genres := make([]types.TMDbGenre, 0, len(movieGenres.Genres))
	for _, genre := range movieGenres.Genres {
		genres = append(genres, types.TMDbGenre{ID: genre.ID, Name: genre.Name})
	}
	tvGenres, err := l.api.GetTvGenres(l.options)
	if err != nil {
		return fmt.Errorf("failed to save tv genres: %s", err.Error())
	}
	for _, genre := range tvGenres.Genres {
		genres = append(genres, types.TMDbGenre{ID: genre.ID, Name: genre.Name})
	}
	err = l.storer.InsertTMDbGenres(ctx, genres)
	if err != nil {
		return fmt.Errorf("failed to save tv genres: %s", err.Error())
	}

	err1Ch := make(chan error)
	err2Ch := make(chan error)
	go func(errCh chan error) {
		l.log.Info("start discover and save movies")
		err := l.DiscoverAndSave(ctx, Movie)
		err1Ch <- err
	}(err1Ch)
	go func(errCh chan error) {
		l.log.Info("start discover and save tvs")
		err := l.DiscoverAndSave(ctx, TV)
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

func (l *TMDbLoader) DiscoverAndSave(ctx context.Context, dt DiscoverType) error {
	fromPage := 1
	toPage := 500
	for page := fromPage; page <= toPage; page++ {
		l.options["page"] = fmt.Sprintf("%d", page)

		switch dt {
		case Movie:
			res, err := l.api.DiscoverMovie(l.options)
			if err != nil {
				l.log.Error("failed to discover movies", "error", err.Error())
				return err
			}
			l.log.Info("discover movies success", "page", page, "total_pages", res.TotalPages, "total_results", res.TotalResults)
			movies := make([]types.TMDbMovie, 0, len(res.Results))
			moviesGenresMap := make(map[int][]int32, len(res.Results))
			for _, movie := range res.Results {
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

				moviesGenresMap[movie.ID] = movie.GenreIDs
			}

			err = l.storer.InsertTMDbMovies(ctx, movies)
			if err != nil {
				l.log.Error("failed to insert movies", "error", err.Error(), "page", page, "movies", movies)
				return err
			}

			for movieID, genreIDs := range moviesGenresMap {
				err = l.storer.InsertTMDbMoviesGenres(ctx, movieID, genreIDs)
				if err != nil {
					l.log.Error("failed to insert movies genres", "error", err.Error(), "movieID", movieID, "genreIDs", genreIDs)
					return err
				}
			}
		case TV:
			res, err := l.api.DiscoverTV(l.options)
			if err != nil {
				l.log.Error("failed to discover TVs", "error", err.Error())
				return err
			}
			l.log.Info("discover TVs success", "page", page, "total_pages", res.TotalPages, "total_results", res.TotalResults)
			tvs := make([]types.TMDbTV, 0, len(res.Results))
			tvsGenresMap := make(map[int][]int32, len(res.Results))
			for _, tv := range res.Results {
				tvTitle := tv.Name
				if tv.Name == "" {
					tvTitle = tv.OriginalName
				}
				tvs = append(tvs, types.TMDbTV{
					ID:          tv.ID,
					Title:       tvTitle,
					Overview:    tv.Overview,
					Popularity:  tv.Popularity,
					PosterPath:  tv.PosterPath,
					VoteAverage: tv.VoteAverage,
					VoteCount:   tv.VoteCount,
				})
				tvsGenresMap[tv.ID] = tv.GenreIDs
			}

			err = l.storer.InsertTMDbTVs(ctx, tvs)
			if err != nil {
				l.log.Error("failed to insert tvs", "error", err.Error(), "page", page, "tvs", tvs)
				return err
			}

			for tvID, genreIDs := range tvsGenresMap {
				err = l.storer.InsertTMDbTVsGenres(ctx, tvID, genreIDs)
				if err != nil {
					l.log.Error("failed to insert tvs genres", "error", err.Error(), "tvID", tvID, "genreIDs", genreIDs)
					return err
				}
			}
		}
	}
	return nil
}
