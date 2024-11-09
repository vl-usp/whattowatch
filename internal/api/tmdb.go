package api

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"whattowatch/internal/api/cache"
	"whattowatch/internal/config"
	"whattowatch/internal/types"
	"whattowatch/internal/utils"

	tmdb "github.com/cyruzin/golang-tmdb"
	"golang.org/x/sync/errgroup"
)

type (
	TMDbApi struct {
		client *tmdb.Client
		cache  *cache.Cache
		opts   map[string]string

		cfg *config.Config
		log *slog.Logger
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

const recomendationDateFrom = "2012-01-01"
const emptyImageUrl = "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcRRT-fEjKFv3PMMg47-olkMSqEDqD42C7ZAsg&s"
const workers = 5

func New(cfg *config.Config, log *slog.Logger) (*TMDbApi, error) {
	opts := make(map[string]string)
	opts["language"] = "ru-RU"

	c, err := tmdb.Init(cfg.Tokens.TMDb)
	if err != nil {
		return nil, err
	}

	api := &TMDbApi{
		client: c,
		opts:   opts,

		cfg: cfg,
		log: log.With("pkg", "api"),
	}

	api.initCache()
	return api, nil
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
			a.cache.Genres.Set(genre.ID, genre.Name)
		}
		return nil
	})

	g.Go(func() error {
		tvs, err := a.client.GetGenreTVList(a.opts)
		if err != nil {
			return err
		}
		for _, genre := range tvs.Genres {
			a.cache.Genres.Set(genre.ID, genre.Name)
		}

		return nil
	})

	err := g.Wait()

	a.log.Debug("genres loaded", "genres", a.cache.Genres)

	return err
}

func (a *TMDbApi) getGenresByIDs(ids []int64) types.Genres {
	genres := make(types.Genres, 0, len(ids))
	for _, id := range ids {
		if g, ok := a.cache.Genres.Get(id); ok {
			genres = append(genres, types.Genre{
				ID:   id,
				Name: g,
			})
		}
	}

	return genres
}

func (a *TMDbApi) GetMovie(ctx context.Context, id int) (types.ContentItem, error) {
	log := a.log.With("method", "GetMovie", "id", id)
	m, err := a.client.GetMovieDetails(id, a.opts)
	if err != nil {
		return types.ContentItem{}, err
	}
	log.Info("got movie details", "id", m.ID, "title", m.Title)

	rd, err := utils.GetReleaseDate(m.ReleaseDate)
	if err != nil {
		return types.ContentItem{}, err
	}

	genres := make(types.Genres, 0, len(m.Genres))
	for _, genre := range m.Genres {
		genres = append(genres, types.Genre{ID: genre.ID, Name: genre.Name})
	}

	return types.ContentItem{
		ID:          m.ID,
		ContentType: types.Movie,
		Title:       m.Title,
		Overview:    m.Overview,
		Popularity:  m.Popularity,
		PosterPath:  a.cfg.Urls.TMDbImageUrl + m.PosterPath,
		ReleaseDate: rd,
		VoteAverage: m.VoteAverage,
		VoteCount:   m.VoteCount,
		Genres:      genres,
	}, nil
}

func (a *TMDbApi) GetMovies(ctx context.Context, ids []int64) (types.Content, error) {
	log := a.log.With("method", "GetMovies")
	content := make(types.Content, 0, len(ids))

	jobCh := make(chan int64, len(ids))
	movieCh := make(chan contentItem, len(ids))
	defer close(movieCh)

	log.Info("start working pool", "ids", ids)
	for i := 0; i < workers; i++ {
		go func(id int, jobCh <-chan int64, movieCh chan<- contentItem) {
			for job := range jobCh {
				m, err := a.GetMovie(ctx, int(job))
				movieCh <- contentItem{
					contentItem: m,
					err:         err,
				}
			}

		}(i, jobCh, movieCh)
	}

	for _, id := range ids {
		jobCh <- id
	}
	defer close(jobCh)

	for i := 0; i < len(ids); i++ {
		m := <-movieCh
		if m.err != nil {
			return nil, m.err
		}
		content = append(content, m.contentItem)
	}

	return content, nil
}

func (a *TMDbApi) GetTV(ctx context.Context, id int) (types.ContentItem, error) {
	log := a.log.With("method", "GetTV", "id", id)
	tv, err := a.client.GetTVDetails(id, a.opts)
	if err != nil {
		return types.ContentItem{}, err
	}
	log.Info("got tv details", "id", tv.ID, "title", tv.Name)

	rd, err := utils.GetReleaseDate(tv.FirstAirDate)
	if err != nil {
		return types.ContentItem{}, err
	}

	genres := make(types.Genres, 0, len(tv.Genres))
	for _, genre := range tv.Genres {
		genres = append(genres, types.Genre{ID: genre.ID, Name: genre.Name})
	}

	return types.ContentItem{
		ID:          tv.ID,
		ContentType: types.TV,
		Title:       tv.Name,
		Overview:    tv.Overview,
		Popularity:  tv.Popularity,
		PosterPath:  a.cfg.Urls.TMDbImageUrl + tv.PosterPath,
		ReleaseDate: rd,
		VoteAverage: tv.VoteAverage,
		VoteCount:   tv.VoteCount,
		Genres:      genres,
	}, nil
}

func (a *TMDbApi) GetTVs(ctx context.Context, ids []int64) (types.Content, error) {
	log := a.log.With("method", "GetTVs")
	content := make(types.Content, 0, len(ids))

	jobCh := make(chan int64, len(ids))
	tvCh := make(chan contentItem, len(ids))
	defer close(tvCh)

	log.Info("start working pool", "ids", ids)
	for i := 0; i < workers; i++ {
		go func(id int, jobCh <-chan int64, tvCh chan<- contentItem) {
			for job := range jobCh {
				m, err := a.GetTV(ctx, int(job))
				tvCh <- contentItem{
					contentItem: m,
					err:         err,
				}
			}

		}(i, jobCh, tvCh)
	}

	for _, id := range ids {
		jobCh <- id
	}
	defer close(jobCh)

	for i := 0; i < len(ids); i++ {
		m := <-tvCh
		if m.err != nil {
			return nil, m.err
		}
		content = append(content, m.contentItem)
	}

	return content, nil
}

func (a *TMDbApi) GetMoviePopular(ctx context.Context, page int) (types.Content, error) {
	log := a.log.With("fn", "GetMoviePopular", "page", page)
	a.opts["page"] = fmt.Sprintf("%d", page)
	defer delete(a.opts, "page")

	m, err := a.client.GetMoviePopular(a.opts)
	if err != nil {
		return nil, err
	}
	log.Info("got movie popular", "count", len(m.MoviePopularResults.Results))

	res := make(types.Content, 0, len(m.MoviePopularResults.Results))
	for _, v := range m.MoviePopularResults.Results {
		rd, err := utils.GetReleaseDate(v.ReleaseDate)
		if err != nil {
			continue
		}

		title := v.Title
		if title == "" {
			title = v.OriginalTitle
		}

		res = append(res, types.ContentItem{
			ID:          v.ID,
			ContentType: types.Movie,
			Title:       title,
			Overview:    v.Overview,
			Popularity:  v.Popularity,
			PosterPath:  a.cfg.Urls.TMDbImageUrl + v.PosterPath,
			ReleaseDate: rd,
			VoteAverage: v.VoteAverage,
			VoteCount:   v.VoteCount,
		})
	}

	return res, nil
}

func (a *TMDbApi) GetTVPopular(ctx context.Context, page int) (types.Content, error) {
	log := a.log.With("fn", "GetTVPopular", "page", page)

	a.opts["page"] = fmt.Sprintf("%d", page)
	defer delete(a.opts, "page")

	m, err := a.client.GetTVPopular(a.opts)
	if err != nil {
		return nil, err
	}
	log.Info("got tv popular", "count", len(m.Results))

	res := make(types.Content, 0, len(m.Results))
	for _, v := range m.Results {
		rd, err := utils.GetReleaseDate(v.FirstAirDate)
		if err != nil {
			continue
		}

		title := v.Name
		if title == "" {
			title = v.OriginalName
		}

		res = append(res, types.ContentItem{
			ID:          v.ID,
			ContentType: types.TV,
			Title:       title,
			Overview:    v.Overview,
			Popularity:  v.Popularity,
			PosterPath:  a.cfg.Urls.TMDbImageUrl + v.PosterPath,
			ReleaseDate: rd,
			VoteAverage: v.VoteAverage,
			VoteCount:   v.VoteCount,
		})
	}

	return res, nil
}

func (a *TMDbApi) GetMovieTop(ctx context.Context, page int) (types.Content, error) {
	log := a.log.With("fn", "GetMovieTop", "page", page)

	a.opts["page"] = fmt.Sprintf("%d", page)
	defer delete(a.opts, "page")

	m, err := a.client.GetMovieTopRated(a.opts)
	if err != nil {
		return nil, err
	}
	log.Info("got movie top rated", "count", len(m.Results))

	res := make(types.Content, 0, len(m.Results))
	for _, v := range m.Results {
		rd, err := utils.GetReleaseDate(v.ReleaseDate)
		if err != nil {
			continue
		}

		title := v.Title
		if title == "" {
			title = v.OriginalTitle
		}

		res = append(res, types.ContentItem{
			ID:          v.ID,
			ContentType: types.Movie,
			Title:       title,
			Overview:    v.Overview,
			Popularity:  v.Popularity,
			PosterPath:  a.cfg.Urls.TMDbImageUrl + v.PosterPath,
			ReleaseDate: rd,
			VoteAverage: v.VoteAverage,
			VoteCount:   v.VoteCount,
		})
	}

	return res, nil
}

func (a *TMDbApi) GetTVTop(ctx context.Context, page int) (types.Content, error) {
	log := a.log.With("fn", "GetTVTop", "page", page)

	a.opts["page"] = fmt.Sprintf("%d", page)
	defer delete(a.opts, "page")

	m, err := a.client.GetTVTopRated(a.opts)
	if err != nil {
		return nil, err
	}
	log.Info("got tv top rated", "count", len(m.Results))

	res := make(types.Content, 0, len(m.Results))
	for _, v := range m.Results {
		rd, err := utils.GetReleaseDate(v.FirstAirDate)
		if err != nil {
			continue
		}

		title := v.Name
		if title == "" {
			title = v.OriginalName
		}

		res = append(res, types.ContentItem{
			ID:          v.ID,
			ContentType: types.TV,
			Title:       title,
			Overview:    v.Overview,
			Popularity:  v.Popularity,
			PosterPath:  a.cfg.Urls.TMDbImageUrl + v.PosterPath,
			ReleaseDate: rd,
			VoteAverage: v.VoteAverage,
			VoteCount:   v.VoteCount,
		})
	}

	return res, nil
}

func (a *TMDbApi) GetMovieRecommendations(ctx context.Context, ids []int64) (types.Content, error) {
	log := a.log.With("method", "GetMovieRecomendations")

	jobCh := make(chan int64, len(ids))
	movieCh := make(chan content, len(ids))
	defer close(movieCh)

	for i := 0; i < workers; i++ {
		go func(id int, jobCh <-chan int64, movieCh chan<- content) {
			for job := range jobCh {
				res, err := a.client.GetMovieRecommendations(int(job), a.opts)
				log.Info("request to TMDb", "worker_id", id, "movie_id", job)
				if err != nil {
					log.Error("request error", "id", id, "error", err.Error())
					movieCh <- content{
						err: err,
					}
				}

				c := make(types.Content, 0, len(res.Results))
				for _, v := range res.Results {
					if v.Overview == "" {
						log.Warn("empty overview", "id", v.ID)
						continue
					}

					rd, err := utils.GetReleaseDate(v.ReleaseDate)
					if err != nil {
						log.Error("get release date error", "id", v.ID, "error", err.Error())
						continue
					}

					filterTime, err := time.Parse("2006-01-02", recomendationDateFrom)
					if err != nil {
						log.Error("parse filter date error", "error", err.Error())
						continue
					}

					if rd.Time.Before(filterTime) {
						continue
					}

					c = append(c, types.ContentItem{
						ID:          v.ID,
						ContentType: types.Movie,
						Title:       v.Title,
						Overview:    v.Overview,
						Popularity:  v.Popularity,
						PosterPath:  a.cfg.Urls.TMDbImageUrl + v.PosterPath,
						ReleaseDate: rd,
						VoteAverage: v.VoteAverage,
						VoteCount:   v.VoteCount,
						Genres:      a.getGenresByIDs(v.GenreIDs),
					})
				}

				movieCh <- content{
					content: c,
					err:     err,
				}
			}

		}(i, jobCh, movieCh)
	}

	for _, id := range ids {
		jobCh <- id
	}
	defer close(jobCh)

	result := make(types.Content, 0)
	for i := 0; i < len(ids); i++ {
		m := <-movieCh
		if m.err != nil {
			return nil, m.err
		}
		result = append(result, m.content...)
	}

	return result, nil
}

func (a *TMDbApi) GetTVRecommendations(ctx context.Context, ids []int64) (types.Content, error) {
	log := a.log.With("method", "GetTVRecomendations")

	jobCh := make(chan int64, len(ids))
	tvCh := make(chan content, len(ids))
	defer close(tvCh)

	for i := 0; i < workers; i++ {
		go func(id int, jobCh <-chan int64, tvCh chan<- content) {
			for job := range jobCh {
				res, err := a.client.GetTVRecommendations(int(job), a.opts)
				log.Info("request to TMDb", "worker_id", id, "tv_id", job)
				if err != nil {
					log.Error("request error", "id", id, "error", err.Error())
					tvCh <- content{
						err: err,
					}
				}

				c := make(types.Content, 0, len(res.Results))
				for _, v := range res.Results {
					if v.Overview == "" {
						log.Warn("empty overview", "id", v.ID)
						continue
					}

					rd, err := utils.GetReleaseDate(v.FirstAirDate)
					if err != nil {
						log.Error("get release date error", "id", v.ID, "error", err.Error())
						continue
					}

					filterTime, err := time.Parse("2006-01-02", recomendationDateFrom)
					if err != nil {
						log.Error("parse filter date error", "error", err.Error())
						continue
					}

					if rd.Time.Before(filterTime) {
						continue
					}

					c = append(c, types.ContentItem{
						ID:          v.ID,
						ContentType: types.TV,
						Title:       v.Name,
						Overview:    v.Overview,
						Popularity:  v.Popularity,
						PosterPath:  a.cfg.Urls.TMDbImageUrl + v.PosterPath,
						ReleaseDate: rd,
						VoteAverage: v.VoteAverage,
						VoteCount:   v.VoteCount,
						Genres:      a.getGenresByIDs(v.GenreIDs),
					})
				}

				tvCh <- content{
					content: c,
					err:     err,
				}
			}

		}(i, jobCh, tvCh)
	}

	for _, id := range ids {
		jobCh <- id
	}
	defer close(jobCh)

	result := make(types.Content, 0)
	for i := 0; i < len(ids); i++ {
		m := <-tvCh
		if m.err != nil {
			return nil, m.err
		}
		result = append(result, m.content...)
	}

	return result, nil
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

func (a *TMDbApi) searchMovieByTitle(_ context.Context, titles []string) (types.Content, error) {
	log := a.log.With("method", "SearchMovieByTitle")

	jobCh := make(chan string, len(titles))
	movieCh := make(chan content, len(titles))
	defer close(movieCh)

	for i := 0; i < workers; i++ {
		go func(id int, jobCh <-chan string, movieCh chan<- content) {
			for job := range jobCh {
				res, err := a.client.GetSearchMovies(job, a.opts)
				log.Info("request to TMDb", "worker_id", id, "title", job)
				if err != nil {
					log.Error("request error", "id", id, "error", err.Error())
					movieCh <- content{
						err: err,
					}
				}

				c := make(types.Content, 0, len(res.Results))
				for _, v := range res.Results {
					if v.Title != job {
						continue
					}

					rd, err := utils.GetReleaseDate(v.ReleaseDate)
					if err != nil {
						log.Warn("get release date error", "id", v.ID, "error", err.Error())
						continue
					}

					poster := a.cfg.Urls.TMDbImageUrl + v.PosterPath
					if v.PosterPath == "" {
						log.Warn("empty poster path", "id", v.ID)
						poster = emptyImageUrl
					}

					c = append(c, types.ContentItem{
						ID:          v.ID,
						ContentType: types.Movie,
						Title:       v.Title,
						Overview:    v.Overview,
						Popularity:  v.Popularity,
						PosterPath:  poster,
						ReleaseDate: rd,
						VoteAverage: v.VoteAverage,
						VoteCount:   v.VoteCount,
					})
				}

				movieCh <- content{
					content: c,
					err:     err,
				}
			}

		}(i, jobCh, movieCh)
	}

	for _, title := range titles {
		jobCh <- title
	}
	defer close(jobCh)

	result := make(types.Content, 0)
	for i := 0; i < len(titles); i++ {
		m := <-movieCh
		if m.err != nil {
			return nil, m.err
		}
		result = append(result, m.content...)
	}

	return result, nil
}

func (a *TMDbApi) searchTVByTitle(_ context.Context, titles []string) (types.Content, error) {
	log := a.log.With("method", "SearchTVByTitle")

	jobCh := make(chan string, len(titles))
	tvCh := make(chan content, len(titles))
	defer close(tvCh)

	for i := 0; i < workers; i++ {
		go func(id int, jobCh <-chan string, movieCh chan<- content) {
			for job := range jobCh {
				res, err := a.client.GetSearchTVShow(job, a.opts)
				log.Info("request to TMDb", "worker_id", id, "title", job)
				if err != nil {
					log.Error("request error", "id", id, "error", err.Error())
					movieCh <- content{
						err: err,
					}
				}

				c := make(types.Content, 0, len(res.Results))
				for _, v := range res.Results {
					if v.Name != job {
						continue
					}

					rd, err := utils.GetReleaseDate(v.FirstAirDate)
					if err != nil {
						log.Warn("get release date error", "id", v.ID, "error", err.Error())
						continue
					}

					poster := a.cfg.Urls.TMDbImageUrl + v.PosterPath
					if v.PosterPath == "" {
						log.Warn("empty poster path", "id", v.ID)
						poster = emptyImageUrl
					}

					c = append(c, types.ContentItem{
						ID:          v.ID,
						ContentType: types.TV,
						Title:       v.Name,
						Overview:    v.Overview,
						Popularity:  v.Popularity,
						PosterPath:  poster,
						ReleaseDate: rd,
						VoteAverage: v.VoteAverage,
						VoteCount:   v.VoteCount,
					})
				}

				tvCh <- content{
					content: c,
					err:     err,
				}
			}

		}(i, jobCh, tvCh)
	}

	for _, title := range titles {
		jobCh <- title
	}
	defer close(jobCh)

	result := make(types.Content, 0)
	for i := 0; i < len(titles); i++ {
		m := <-tvCh
		if m.err != nil {
			return nil, m.err
		}
		result = append(result, m.content...)
	}

	return result, nil
}
