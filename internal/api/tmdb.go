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

type TMDbApi struct {
	client *tmdb.Client
	opts   map[string]string
	cache  *cache.Cache

	cfg *config.Config
	log *slog.Logger
}

type content struct {
	content types.Content
	err     error
}

var recomendationDateFrom = "2012-01-01"

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
	m, err := a.client.GetMovieDetails(id, a.opts)
	if err != nil {
		return types.ContentItem{}, err
	}

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

// TODO: make concurrent requests
func (a *TMDbApi) GetMovies(ctx context.Context, ids []int64) (types.Content, error) {
	content := make(types.Content, 0, len(ids))

	for i := 0; i < len(ids); i++ {
		id := int(ids[i])
		m, err := a.GetMovie(ctx, id)
		if err != nil {
			return nil, err
		}
		content = append(content, m)
	}

	return content, nil
}

func (a *TMDbApi) GetTV(ctx context.Context, id int) (types.ContentItem, error) {
	m, err := a.client.GetTVDetails(id, a.opts)
	if err != nil {
		return types.ContentItem{}, err
	}

	rd, err := utils.GetReleaseDate(m.FirstAirDate)
	if err != nil {
		return types.ContentItem{}, err
	}

	genres := make(types.Genres, 0, len(m.Genres))
	for _, genre := range m.Genres {
		genres = append(genres, types.Genre{ID: genre.ID, Name: genre.Name})
	}

	return types.ContentItem{
		ID:          m.ID,
		ContentType: types.TV,
		Title:       m.Name,
		Overview:    m.Overview,
		Popularity:  m.Popularity,
		PosterPath:  a.cfg.Urls.TMDbImageUrl + m.PosterPath,
		ReleaseDate: rd,
		VoteAverage: m.VoteAverage,
		VoteCount:   m.VoteCount,
		Genres:      genres,
	}, nil
}

// TODO: make concurrent requests
func (a *TMDbApi) GetTVs(ctx context.Context, ids []int64) (types.Content, error) {
	content := make(types.Content, 0, len(ids))

	for i := 0; i < len(ids); i++ {
		id := int(ids[i])
		m, err := a.GetTV(ctx, id)
		if err != nil {
			return nil, err
		}
		content = append(content, m)
	}

	return content, nil
}

func (a *TMDbApi) GetMoviePopular(ctx context.Context, page int) (types.Content, error) {
	a.opts["page"] = fmt.Sprintf("%d", page)
	defer delete(a.opts, "page")

	m, err := a.client.GetMoviePopular(a.opts)
	if err != nil {
		return nil, err
	}

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
	a.opts["page"] = fmt.Sprintf("%d", page)
	defer delete(a.opts, "page")

	m, err := a.client.GetTVPopular(a.opts)
	if err != nil {
		return nil, err
	}

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
	a.opts["page"] = fmt.Sprintf("%d", page)
	defer delete(a.opts, "page")

	m, err := a.client.GetMovieTopRated(a.opts)
	if err != nil {
		return nil, err
	}

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
	a.opts["page"] = fmt.Sprintf("%d", page)
	defer delete(a.opts, "page")

	m, err := a.client.GetTVTopRated(a.opts)
	if err != nil {
		return nil, err
	}

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

// TODO: make concurrent requests
func (a *TMDbApi) GetMovieRecommendations(ctx context.Context, ids []int64) (types.Content, error) {
	log := a.log.With("method", "GetMovieRecomendations")

	content := make(types.Content, 0, len(ids))

	for i := 0; i < len(ids); i++ {
		id := int(ids[i])
		a.opts["page"] = "1"
		res, err := a.client.GetMovieRecommendations(id, a.opts)
		log.Info("request to TMDb", "id", id)
		if err != nil {
			return nil, err
		}

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

			content = append(content, types.ContentItem{
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
	}

	return content, nil
}

// TODO: make concurrent requests
func (a *TMDbApi) GetTVRecommendations(ctx context.Context, ids []int64) (types.Content, error) {
	log := a.log.With("method", "GetTVRecomendations")

	content := make(types.Content, 0, len(ids))

	for i := 0; i < len(ids); i++ {
		id := int(ids[i])
		a.opts["page"] = "1"
		res, err := a.client.GetTVRecommendations(id, a.opts)
		log.Info("request to TMDb", "id", id)
		if err != nil {
			return nil, err
		}

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

			content = append(content, types.ContentItem{
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
	}

	return content, nil
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

// TODO: make concurrent requests
func (a *TMDbApi) searchMovieByTitle(_ context.Context, titles []string) (types.Content, error) {
	log := a.log.With("method", "SearchMovieByTitle")

	content := make(types.Content, 0, len(titles))
	for _, title := range titles {
		res, err := a.client.GetSearchMovies(title, a.opts)
		log.Info("request to TMDb", "title", title)
		if err != nil {
			log.Error("request error", "error", err.Error())
			return nil, err
		}

		for _, v := range res.Results {
			if v.Title != title {
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
				poster = "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcRRT-fEjKFv3PMMg47-olkMSqEDqD42C7ZAsg&s"
			}

			content = append(content, types.ContentItem{
				ID:          v.ID,
				ContentType: types.Movie,
				Title:       title,
				Overview:    v.Overview,
				Popularity:  v.Popularity,
				PosterPath:  poster,
				ReleaseDate: rd,
				VoteAverage: v.VoteAverage,
				VoteCount:   v.VoteCount,
			})
		}

	}

	return content, nil
}

// TODO: make concurrent requests
func (a *TMDbApi) searchTVByTitle(_ context.Context, titles []string) (types.Content, error) {
	log := a.log.With("method", "SearchTVByTitle")

	content := make(types.Content, 0, len(titles))
	for _, title := range titles {
		res, err := a.client.GetSearchTVShow(title, a.opts)
		log.Info("request to TMDb", "title", title)
		if err != nil {
			log.Error("request error", "error", err.Error())
			return nil, err
		}

		for _, v := range res.Results {
			if v.Name != title {
				continue
			}

			poster := a.cfg.Urls.TMDbImageUrl + v.PosterPath
			if v.PosterPath == "" {
				log.Warn("empty poster path", "id", v.ID)
				poster = "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcRRT-fEjKFv3PMMg47-olkMSqEDqD42C7ZAsg&s"
			}

			rd, err := utils.GetReleaseDate(v.FirstAirDate)
			if err != nil {
				log.Warn("get release date error", "id", v.ID, "error", err.Error())
				continue
			}

			content = append(content, types.ContentItem{
				ID:          v.ID,
				ContentType: types.TV,
				Title:       title,
				Overview:    v.Overview,
				Popularity:  v.Popularity,
				PosterPath:  poster,
				ReleaseDate: rd,
				VoteAverage: v.VoteAverage,
				VoteCount:   v.VoteCount,
			})
		}
	}

	return content, nil
}
