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

func NewTMDbLoader(cfg *config.Config, logger *slog.Logger, storer storage.Storer) (*TMDbLoader, error) {
	u, err := url.Parse(cfg.Urls.TMDbApiUrl)
	if err != nil {
		return nil, err
	}

	c, err := tmdbLib.Init(cfg.Tokens.TMDb)
	if err != nil {
		return nil, err
	}
	c.SetClientAutoRetry()

	log := logger.With("pkg", "loader")

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
		genres[genre.ID] = types.Genre{ID: genre.ID, Name: genre.Name}
	}

	tvGenres, err := l.client.GetGenreTVList(l.options)
	if err != nil {
		return fmt.Errorf("failed to get tv genres: %s", err.Error())
	}

	for _, genre := range tvGenres.Genres {
		genres[genre.ID] = types.Genre{ID: genre.ID, Name: genre.Name}
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

		l.log.Info("trying to discover", "page", page, "content type", dt)

		content := make(types.Content, 0)
		switch dt {
		case types.Movie:
			res, err := l.client.GetDiscoverMovie(l.options)
			if err != nil {
				return err
			}
			content = l.convertMovies(res)
		case types.TV:
			res, err := l.client.GetDiscoverTV(l.options)
			if err != nil {
				return err
			}
			content = l.convertTVs(res)
		}

		if len(content) == 0 {
			l.log.Info("discover content empty", "page", page, "content type", dt)
			continue
		}

		l.log.Info("discover content success", "page", page, "content type", dt, "count", len(content))

		err := l.storer.InsertContent(ctx, content)
		if err != nil {
			return err
		}

		l.log.Info("insert content success", "page", page, "content type", dt, "count", len(content))
	}
	return nil
}

func (l *TMDbLoader) convertMovies(movies *tmdbLib.DiscoverMovie) types.Content {
	result := make(types.Content, 0, len(movies.Results))

	for _, movie := range movies.Results {
		// skip movies with empty overview
		if movie.Overview == "" {
			continue
		}

		releaseDate, err := utils.GetReleaseDate(movie.ReleaseDate)
		if err != nil {
			l.log.Error("failed to get release date", "error", err.Error(), "id", movie.ID, "release_date", movie.ReleaseDate)
		}

		genres := make(types.Genres, 0, len(movie.GenreIDs))
		for _, genreID := range movie.GenreIDs {
			genres = append(genres, types.Genre{ID: genreID})
		}

		result = append(result, types.ContentItem{
			ID:          movie.ID,
			ContentType: types.Movie,
			Title:       movie.Title,
			Overview:    movie.Overview,
			Popularity:  movie.Popularity,
			PosterPath:  l.cfg.Urls.TMDbImageUrl + movie.PosterPath,
			ReleaseDate: releaseDate,
			VoteAverage: movie.VoteAverage,
			VoteCount:   movie.VoteCount,
			Genres:      genres,
		})
	}

	return result
}

func (l *TMDbLoader) convertTVs(tvs *tmdbLib.DiscoverTV) types.Content {
	result := make(types.Content, 0, len(tvs.Results))

	for _, tv := range tvs.Results {
		// skip tvs with empty overview
		if tv.Overview == "" {
			continue
		}

		tvTitle := tv.Name
		if tv.Name == "" {
			tvTitle = tv.OriginalName
		}

		releaseDate, err := utils.GetReleaseDate(tv.FirstAirDate)
		if err != nil {
			l.log.Error("failed to get release date", "error", err.Error(), "id", tv.ID, "release_date", tv.FirstAirDate)
		}

		genres := make(types.Genres, 0, len(tv.GenreIDs))
		for _, genreID := range tv.GenreIDs {
			genres = append(genres, types.Genre{ID: genreID})
		}

		result = append(result, types.ContentItem{
			ID:          tv.ID,
			ContentType: types.TV,
			Title:       tvTitle,
			Overview:    tv.Overview,
			Popularity:  tv.Popularity,
			ReleaseDate: releaseDate,
			PosterPath:  l.cfg.Urls.TMDbImageUrl + tv.PosterPath,
			VoteAverage: tv.VoteAverage,
			VoteCount:   tv.VoteCount,
			Genres:      genres,
		})
	}

	return result
}
