package tmdb

import (
	"context"
	"errors"
	"log/slog"
	"whattowatch/internal/api/cache"
	"whattowatch/internal/config"
	"whattowatch/internal/types"

	tmdb "github.com/cyruzin/golang-tmdb"
	"golang.org/x/sync/errgroup"
)

type (
	TMDbApi struct {
		client *tmdb.Client
		cache  *cache.Cache

		cfg *config.Config
		log *slog.Logger

		opts map[string]string
	}

	content struct {
		content types.Content
		err     error
	}

	contentItem struct {
		contentItem types.ContentItem
		err         error
	}
)

const workers = 5

func New(cfg *config.Config, log *slog.Logger) (*TMDbApi, error) {
	opts := make(map[string]string)
	opts["language"] = "ru-RU"

	c, err := tmdb.Init(cfg.Tokens.TMDb)
	if err != nil {
		return nil, err
	}
	c.SetAlternateBaseURL()
	c.SetClientAutoRetry()

	api := &TMDbApi{
		client: c,
		opts:   opts,

		cfg: cfg,
		log: log.With("pkg", "api"),
	}

	api.initCache()
	return api, nil
}

func (a *TMDbApi) GetContent(ctx context.Context, contentType types.ContentType, ids []int64) (types.Content, error) {
	log := a.log.With("fn", "GetContent", "content_type", contentType, "ids", ids)
	log.Debug("func start log")

	jobCh := make(chan int64, len(ids))
	contentItemCh := make(chan contentItem, len(ids))
	defer close(contentItemCh)

	log.Debug("start working pool", "ids", ids)
	for i := 0; i < workers; i++ {
		go func(id int, jobCh <-chan int64, contentItemCh chan<- contentItem) {
			for job := range jobCh {

				var ci types.ContentItem
				var err error

				switch contentType {
				case types.Movie:
					ci, err = a.GetMovie(ctx, int(job))
				case types.TV:
					ci, err = a.GetTV(ctx, int(job))
				}
				contentItemCh <- contentItem{
					contentItem: ci,
					err:         err,
				}
			}

		}(i, jobCh, contentItemCh)
	}

	go func() {
		for _, id := range ids {
			jobCh <- id
		}
		close(jobCh)
	}()

	content := make(types.Content, 0, len(ids))
	for i := 0; i < len(ids); i++ {
		m := <-contentItemCh
		if m.err != nil {
			return nil, m.err
		}
		content = append(content, m.contentItem)
	}

	return content, nil
}

func (a *TMDbApi) GetRecommendations(ctx context.Context, contentType types.ContentType, ids []int64) (types.Content, error) {
	switch contentType {
	case types.Movie:
		return a.GetMovieRecommendations(ctx, ids)
	case types.TV:
		return a.GetTVRecommendations(ctx, ids)
	}

	return nil, errors.New("unknown content type")
}

func (a *TMDbApi) GetGenres(ctx context.Context, contentType types.ContentType) (types.Genres, error) {
	log := a.log.With("fn", "GetGenres", "content_type", contentType)
	log.Debug("func start log")

	var genresMap map[int64]string
	switch contentType {
	case types.Movie:
		genresMap = a.cache.Genres.Movie.GetAll()
	case types.TV:
		genresMap = a.cache.Genres.TV.GetAll()
	}

	res := make(types.Genres, 0, len(genresMap))
	for k, v := range genresMap {
		res = append(res, types.Genre{
			ID:   k,
			Name: v,
		})
	}

	return res, nil
}

func (a *TMDbApi) SearchByTitles(ctx context.Context, titles []string) (types.Content, error) {
	moviesCh := make(chan content)
	tvsCh := make(chan content)

	go func(moviesCh chan content) {
		movies, err := a.searchMovieByTitle(ctx, titles)
		moviesCh <- content{content: movies, err: err}
	}(moviesCh)

	go func(tvsCh chan content) {
		tvs, err := a.searchTVByTitle(ctx, titles)
		tvsCh <- content{content: tvs, err: err}
	}(tvsCh)

	movies := <-moviesCh
	if movies.err != nil {
		return nil, movies.err
	}
	tvs := <-tvsCh
	if tvs.err != nil {
		return nil, tvs.err
	}

	return append(movies.content, tvs.content...), nil
}

func (a *TMDbApi) getOpts() map[string]string {
	optsCopy := make(map[string]string, len(a.opts))
	for k, v := range a.opts {
		optsCopy[k] = v
	}
	return optsCopy
}

func (a *TMDbApi) initCache() error {
	a.cache = cache.New()

	g, _ := errgroup.WithContext(context.Background())
	g.Go(func() error {
		genres, err := a.client.GetGenreMovieList(a.opts)
		if err != nil {
			return err
		}
		for _, genre := range genres.Genres {
			a.cache.Genres.Movie.Set(genre.ID, genre.Name)
		}
		return nil
	})

	g.Go(func() error {
		tvs, err := a.client.GetGenreTVList(a.opts)
		if err != nil {
			return err
		}
		for _, genre := range tvs.Genres {
			a.cache.Genres.TV.Set(genre.ID, genre.Name)
		}

		return nil
	})

	err := g.Wait()

	a.log.Debug("genres loaded", "movies count", len(a.cache.Genres.Movie.GetAll()), "tvs count", len(a.cache.Genres.TV.GetAll()))

	return err
}
