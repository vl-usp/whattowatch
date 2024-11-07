package api

import (
	"context"
	"fmt"
	"log/slog"
	"whattowatch/internal/cache"
	"whattowatch/internal/config"
	"whattowatch/internal/storage"
	"whattowatch/internal/types"
	"whattowatch/internal/utils"

	tmdb "github.com/cyruzin/golang-tmdb"
	"golang.org/x/sync/errgroup"
)

type TMDbApi struct {
	client *tmdb.Client
	opts   map[string]string
	storer storage.Storer
	cache  *cache.Genres

	cfg *config.Config
	log *slog.Logger
}

func New(cfg *config.Config, storer storage.Storer, log *slog.Logger) (*TMDbApi, error) {
	opts := make(map[string]string)
	opts["language"] = "ru-RU"

	c, err := tmdb.Init(cfg.Tokens.TMDb)
	if err != nil {
		return nil, err
	}

	api := &TMDbApi{
		client: c,
		opts:   opts,
		storer: storer,

		cfg: cfg,
		log: log.With("pkg", "api"),
	}

	api.InitGenres()
	return api, nil
}

func (a *TMDbApi) InitGenres() error {
	a.cache = cache.NewGenres()

	g, _ := errgroup.WithContext(context.Background())
	g.Go(func() error {
		genres, err := a.client.GetGenreMovieList(a.opts)
		if err != nil {
			return err
		}
		for _, genre := range genres.Genres {
			a.cache.SetGenre(int(genre.ID), genre.Name)
		}
		return nil
	})

	g.Go(func() error {
		tvs, err := a.client.GetGenreTVList(a.opts)
		if err != nil {
			return err
		}
		for _, genre := range tvs.Genres {
			a.cache.SetGenre(int(genre.ID), genre.Name)
		}

		return nil
	})

	err := g.Wait()
	return err
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

func (a *TMDbApi) GetMoviesPopular(ctx context.Context, page int) (types.Content, error) {
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

		genres := make(types.Genres, 0, len(v.Genres))
		for _, g := range v.Genres {
			genres = append(genres, types.Genre{
				ID:   g.ID,
				Name: g.Name,
			})
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
			Genres:      genres,
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

		genres := make(types.Genres, 0, len(v.GenreIDs))
		for _, id := range v.GenreIDs {
			if g, ok := a.cache.GetGenre(int(id)); ok {
				genres = append(genres, types.Genre{
					ID:   id,
					Name: g,
				})
			}
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
			Genres:      genres,
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

		genres := make(types.Genres, 0, len(v.Genres))
		for _, g := range v.Genres {
			genres = append(genres, types.Genre{
				ID:   g.ID,
				Name: g.Name,
			})
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
			Genres:      genres,
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

		genres := make(types.Genres, 0, len(v.GenreIDs))
		for _, id := range v.GenreIDs {
			if g, ok := a.cache.GetGenre(int(id)); ok {
				genres = append(genres, types.Genre{
					ID:   id,
					Name: g,
				})
			}
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
			Genres:      genres,
		})
	}

	return res, nil
}
