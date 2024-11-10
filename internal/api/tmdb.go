package api

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
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
	c.SetClientAutoRetry()
	// c.SetAlternateBaseURL()

	api := &TMDbApi{
		client: c,
		opts:   opts,

		cfg: cfg,
		log: log.With("pkg", "api"),
	}

	// err = api.testConnection()
	// if err != nil {
	// 	return nil, err
	// }

	api.initCache()
	return api, nil
}

func (a *TMDbApi) testConnection() error {
	err := utils.PingHost("google.com", 443)
	if err != nil {
		return fmt.Errorf("ping google.com error: %s", err.Error())
	}
	a.log.Info("ping google.com", "status", "success")

	err = utils.PingHost("api.themoviedb.org", 443)
	if err != nil {
		return fmt.Errorf("ping api.themoviedb.org error: %s", err.Error())
	}
	a.log.Info("ping tmdb", "status", "success")

	return nil
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

	a.log.Debug("genres loaded", "movies", a.cache.Genres.Movie, "tv", a.cache.Genres.TV)

	return err
}

func (a *TMDbApi) getGenresByIDs(ids []int64, contentType types.ContentType) types.Genres {
	genres := make(types.Genres, 0, len(ids))
	for _, id := range ids {
		switch contentType {
		case types.Movie:
			if g, ok := a.cache.Genres.Movie.Get(id); ok {
				genres = append(genres, types.Genre{
					ID:   id,
					Name: g,
				})
			}
		case types.TV:
			if g, ok := a.cache.Genres.TV.Get(id); ok {
				genres = append(genres, types.Genre{
					ID:   id,
					Name: g,
				})
			}
		}
	}

	return genres
}

func (a *TMDbApi) GetMovie(ctx context.Context, id int) (types.ContentItem, error) {
	log := a.log.With("fn", "GetMovie", "id", id)

	optsCopy := make(map[string]string, len(a.opts))
	for k, v := range a.opts {
		optsCopy[k] = v
	}
	optsCopy["append_to_response"] = "videos"

	m, err := a.client.GetMovieDetails(id, optsCopy)
	if err != nil {
		return types.ContentItem{}, err
	}

	log.Debug("got movie details", "id", m.ID, "title", m.Title, "opts", optsCopy)

	rd, err := utils.GetReleaseDate(m.ReleaseDate)
	if err != nil {
		return types.ContentItem{}, err
	}

	genres := make(types.Genres, 0, len(m.Genres))
	for _, genre := range m.Genres {
		genres = append(genres, types.Genre{ID: genre.ID, Name: genre.Name})
	}

	var trailerURL string
	for _, video := range m.Videos.MovieVideos.MovieVideosResults.Results {
		if video.Type != "Trailer" || !(video.Site == "YouTube" || video.Site == "Youtube") || video.Iso3166_1 != "RU" {
			continue
		}

		trailerURL = fmt.Sprintf("https://youtu.be/%s", video.Key)
	}

	return types.ContentItem{
		ID:           m.ID,
		ContentType:  types.Movie,
		Title:        m.Title,
		Overview:     m.Overview,
		Popularity:   m.Popularity,
		PosterPath:   a.cfg.Urls.TMDbImageUrl + m.PosterPath,
		BackdropPath: a.cfg.Urls.TMDbImageUrl + m.BackdropPath,
		ReleaseDate:  rd,
		VoteAverage:  m.VoteAverage,
		VoteCount:    m.VoteCount,
		Genres:       genres,
		Counties:     m.OriginCountry,
		TrailerURL:   trailerURL,
	}, nil
}

func (a *TMDbApi) GetMovies(ctx context.Context, ids []int64) (types.Content, error) {
	log := a.log.With("fn", "GetMovies")
	content := make(types.Content, 0, len(ids))

	jobCh := make(chan int64, len(ids))
	movieCh := make(chan contentItem, len(ids))
	defer close(movieCh)

	log.Debug("start working pool", "ids", ids)
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
	log := a.log.With("fn", "GetTV", "id", id)

	optsCopy := make(map[string]string, len(a.opts))
	for k, v := range a.opts {
		optsCopy[k] = v
	}
	optsCopy["append_to_response"] = "videos"

	tv, err := a.client.GetTVDetails(id, optsCopy)
	if err != nil {
		return types.ContentItem{}, err
	}
	log.Debug("got tv details", "id", tv.ID, "title", tv.Name)

	rd, err := utils.GetReleaseDate(tv.FirstAirDate)
	if err != nil {
		return types.ContentItem{}, err
	}

	genres := make(types.Genres, 0, len(tv.Genres))
	for _, genre := range tv.Genres {
		genres = append(genres, types.Genre{ID: genre.ID, Name: genre.Name})
	}

	var trailerURL string
	for _, video := range tv.Videos.TVVideos.TVVideosResults.Results {
		if video.Type != "Trailer" || !(video.Site == "YouTube" || video.Site == "Youtube") || video.Iso3166_1 != "RU" {
			continue
		}

		trailerURL = fmt.Sprintf("https://youtu.be/%s", video.Key)
	}

	return types.ContentItem{
		ID:           tv.ID,
		ContentType:  types.TV,
		Title:        tv.Name,
		Overview:     tv.Overview,
		Popularity:   tv.Popularity,
		PosterPath:   a.cfg.Urls.TMDbImageUrl + tv.PosterPath,
		BackdropPath: a.cfg.Urls.TMDbImageUrl + tv.BackdropPath,
		ReleaseDate:  rd,
		VoteAverage:  tv.VoteAverage,
		VoteCount:    tv.VoteCount,
		Genres:       genres,
		Counties:     tv.OriginCountry,
		TrailerURL:   trailerURL,
	}, nil
}

func (a *TMDbApi) GetTVs(ctx context.Context, ids []int64) (types.Content, error) {
	log := a.log.With("fn", "GetTVs")
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
			ID:           v.ID,
			ContentType:  types.Movie,
			Title:        title,
			Overview:     v.Overview,
			Popularity:   v.Popularity,
			PosterPath:   a.cfg.Urls.TMDbImageUrl + v.PosterPath,
			BackdropPath: a.cfg.Urls.TMDbImageUrl + v.BackdropPath,
			ReleaseDate:  rd,
			VoteAverage:  v.VoteAverage,
			VoteCount:    v.VoteCount,
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
			ID:           v.ID,
			ContentType:  types.TV,
			Title:        title,
			Overview:     v.Overview,
			Popularity:   v.Popularity,
			PosterPath:   a.cfg.Urls.TMDbImageUrl + v.PosterPath,
			BackdropPath: a.cfg.Urls.TMDbImageUrl + v.BackdropPath,
			ReleaseDate:  rd,
			VoteAverage:  v.VoteAverage,
			VoteCount:    v.VoteCount,
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
			ID:           v.ID,
			ContentType:  types.Movie,
			Title:        title,
			Overview:     v.Overview,
			Popularity:   v.Popularity,
			PosterPath:   a.cfg.Urls.TMDbImageUrl + v.PosterPath,
			BackdropPath: a.cfg.Urls.TMDbImageUrl + v.BackdropPath,
			ReleaseDate:  rd,
			VoteAverage:  v.VoteAverage,
			VoteCount:    v.VoteCount,
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
			ID:           v.ID,
			ContentType:  types.TV,
			Title:        title,
			Overview:     v.Overview,
			Popularity:   v.Popularity,
			PosterPath:   a.cfg.Urls.TMDbImageUrl + v.PosterPath,
			BackdropPath: a.cfg.Urls.TMDbImageUrl + v.BackdropPath,
			ReleaseDate:  rd,
			VoteAverage:  v.VoteAverage,
			VoteCount:    v.VoteCount,
		})
	}

	return res, nil
}

func (a *TMDbApi) GetMovieRecommendations(ctx context.Context, ids []int64) (types.Content, error) {
	log := a.log.With("fn", "GetMovieRecomendations")

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
						ID:           v.ID,
						ContentType:  types.Movie,
						Title:        v.Title,
						Overview:     v.Overview,
						Popularity:   v.Popularity,
						PosterPath:   a.cfg.Urls.TMDbImageUrl + v.PosterPath,
						BackdropPath: a.cfg.Urls.TMDbImageUrl + v.BackdropPath,
						ReleaseDate:  rd,
						VoteAverage:  v.VoteAverage,
						VoteCount:    v.VoteCount,
						Genres:       a.getGenresByIDs(v.GenreIDs, types.Movie),
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
	log := a.log.With("fn", "GetTVRecomendations")

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
						ID:           v.ID,
						ContentType:  types.TV,
						Title:        v.Name,
						Overview:     v.Overview,
						Popularity:   v.Popularity,
						PosterPath:   a.cfg.Urls.TMDbImageUrl + v.PosterPath,
						BackdropPath: a.cfg.Urls.TMDbImageUrl + v.BackdropPath,
						ReleaseDate:  rd,
						VoteAverage:  v.VoteAverage,
						VoteCount:    v.VoteCount,
						Genres:       a.getGenresByIDs(v.GenreIDs, types.Movie),
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
	log := a.log.With("fn", "SearchMovieByTitle")

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

					backdrop := a.cfg.Urls.TMDbImageUrl + v.BackdropPath
					if v.BackdropPath == "" {
						log.Warn("empty backdrop path", "id", v.ID)
						backdrop = emptyImageUrl
					}

					c = append(c, types.ContentItem{
						ID:           v.ID,
						ContentType:  types.Movie,
						Title:        v.Title,
						Overview:     v.Overview,
						Popularity:   v.Popularity,
						PosterPath:   poster,
						BackdropPath: backdrop,
						ReleaseDate:  rd,
						VoteAverage:  v.VoteAverage,
						VoteCount:    v.VoteCount,
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
	log := a.log.With("fn", "SearchTVByTitle")

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

					backdrop := a.cfg.Urls.TMDbImageUrl + v.BackdropPath
					if v.BackdropPath == "" {
						log.Warn("empty backdrop path", "id", v.ID)
						backdrop = emptyImageUrl
					}

					c = append(c, types.ContentItem{
						ID:           v.ID,
						ContentType:  types.TV,
						Title:        v.Name,
						Overview:     v.Overview,
						Popularity:   v.Popularity,
						PosterPath:   poster,
						BackdropPath: backdrop,
						ReleaseDate:  rd,
						VoteAverage:  v.VoteAverage,
						VoteCount:    v.VoteCount,
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

func (a *TMDbApi) GetMoviesByGenre(ctx context.Context, genreIDs []int, page int) (types.Content, error) {
	log := a.log.With("fn", "DiscoverMovies", "page", page, "genres", genreIDs)

	a.opts["page"] = fmt.Sprintf("%d", page)
	a.opts["with_genres"] = strings.Join(utils.IntSliceToStringSlice(genreIDs), ",")
	defer delete(a.opts, "page")
	defer delete(a.opts, "with_genres")

	m, err := a.client.GetDiscoverMovie(a.opts)
	if err != nil {
		return nil, err
	}
	log.Info("got discover movies", "count", len(m.Results))

	res := make(types.Content, 0, len(m.Results))

	for _, v := range m.Results {
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

		backdrop := a.cfg.Urls.TMDbImageUrl + v.BackdropPath
		if v.BackdropPath == "" {
			log.Warn("empty backdrop path", "id", v.ID)
			backdrop = emptyImageUrl
		}

		res = append(res, types.ContentItem{
			ID:           v.ID,
			ContentType:  types.Movie,
			Title:        v.Title,
			Overview:     v.Overview,
			Popularity:   v.Popularity,
			PosterPath:   poster,
			BackdropPath: backdrop,
			ReleaseDate:  rd,
			VoteAverage:  v.VoteAverage,
			VoteCount:    v.VoteCount,
		})
	}

	return res, nil
}

func (a *TMDbApi) GetTVsByGenre(ctx context.Context, genreIDs []int, page int) (types.Content, error) {
	log := a.log.With("fn", "DiscoverTV", "page", page, "genres", genreIDs)

	a.opts["page"] = fmt.Sprintf("%d", page)
	a.opts["with_genres"] = strings.Join(utils.IntSliceToStringSlice(genreIDs), ",")
	defer delete(a.opts, "page")
	defer delete(a.opts, "with_genres")

	m, err := a.client.GetDiscoverTV(a.opts)
	if err != nil {
		return nil, err
	}
	log.Info("got discover tv", "count", len(m.Results))

	res := make(types.Content, 0, len(m.Results))

	for _, v := range m.Results {
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

		backdrop := a.cfg.Urls.TMDbImageUrl + v.BackdropPath
		if v.BackdropPath == "" {
			log.Warn("empty backdrop path", "id", v.ID)
			backdrop = emptyImageUrl
		}

		res = append(res, types.ContentItem{
			ID:           v.ID,
			ContentType:  types.TV,
			Title:        v.Name,
			Overview:     v.Overview,
			Popularity:   v.Popularity,
			PosterPath:   poster,
			BackdropPath: backdrop,
			ReleaseDate:  rd,
			VoteAverage:  v.VoteAverage,
			VoteCount:    v.VoteCount,
		})
	}

	return res, nil
}
