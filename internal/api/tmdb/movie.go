package tmdb

import (
	"context"
	"fmt"
	"strings"
	"time"
	"whattowatch/internal/api/tmdb/converter"
	"whattowatch/internal/types"
	"whattowatch/internal/utils"
)

func (a *TMDbApi) GetMovie(ctx context.Context, id int) (types.ContentItem, error) {
	log := a.log.With("fn", "GetMovie", "id", id)

	opts := a.getOpts()
	opts["append_to_response"] = "videos"

	m, err := a.client.GetMovieDetails(id, opts)
	if err != nil {
		return types.ContentItem{}, err
	}
	log.Debug("got movie details", "id", m.ID, "title", m.Title, "opts", opts)

	rd, err := time.Parse("2006-01-02", m.ReleaseDate)
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

func (a *TMDbApi) GetMoviePopular(ctx context.Context, page int) (types.Content, error) {
	log := a.log.With("fn", "GetMoviePopular", "page", page)

	opts := a.getOpts()
	opts["page"] = fmt.Sprintf("%d", page)

	m, err := a.client.GetMoviePopular(opts)
	if err != nil {
		return nil, err
	}
	log.Info("got movie popular", "count", len(m.MoviePopularResults.Results))

	res := make(types.Content, 0, len(m.MoviePopularResults.Results))
	for _, v := range m.MoviePopularResults.Results {
		mr := converter.MoviePageResult(v)
		ci, err := mr.Convert(a.cfg.Urls.TMDbImageUrl)
		if err != nil {
			log.Warn("movie result convert error", "id", v.ID, "error", err.Error())
			continue
		}

		res = append(res, ci)
	}

	return res, nil
}

func (a *TMDbApi) GetMovieTop(ctx context.Context, page int) (types.Content, error) {
	log := a.log.With("fn", "GetMovieTop", "page", page)

	opts := a.getOpts()
	opts["page"] = fmt.Sprintf("%d", page)

	m, err := a.client.GetMovieTopRated(opts)
	if err != nil {
		return nil, err
	}
	log.Info("got movie top rated", "count", len(m.Results))

	res := make(types.Content, 0, len(m.Results))
	for _, v := range m.Results {
		mr := converter.MoviePageResult(v)
		ci, err := mr.Convert(a.cfg.Urls.TMDbImageUrl)
		if err != nil {
			log.Warn("movie result convert error", "id", v.ID, "error", err.Error())
			continue
		}

		res = append(res, ci)
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
					log.Error("request error", "id", id, "movie_id", job, "error", err.Error())
					movieCh <- content{
						err: err,
					}
				}

				c := make(types.Content, 0, len(res.Results))
				for _, v := range res.Results {
					mr := converter.MovieRecommendationResult(v)
					ci, err := mr.Convert(a.cfg.Urls.TMDbImageUrl)
					if err != nil {
						log.Warn("movie result convert error", "id", v.ID, "error", err.Error())
						continue
					}
					c = append(c, ci)
				}

				movieCh <- content{
					content: c,
					err:     err,
				}
			}

		}(i, jobCh, movieCh)
	}

	go func() {
		for _, id := range ids {
			jobCh <- id
		}
		close(jobCh)
	}()

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

					mr := converter.MovieSearchResult(v)
					ci, err := mr.Convert(a.cfg.Urls.TMDbImageUrl)
					if err != nil {
						log.Warn("movie result convert error", "id", v.ID, "error", err.Error())
						continue
					}
					c = append(c, ci)
				}

				movieCh <- content{
					content: c,
					err:     err,
				}
			}

		}(i, jobCh, movieCh)
	}

	go func() {
		for _, title := range titles {
			jobCh <- title
		}
		close(jobCh)
	}()

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

func (a *TMDbApi) GetMoviesByGenre(ctx context.Context, genreIDs []int, page int) (types.Content, error) {
	log := a.log.With("fn", "DiscoverMovies", "page", page, "genres", genreIDs)

	opts := a.getOpts()
	opts["page"] = fmt.Sprintf("%d", page)
	opts["with_genres"] = strings.Join(utils.IntSliceToStringSlice(genreIDs), ",")

	m, err := a.client.GetDiscoverMovie(opts)
	if err != nil {
		return nil, err
	}
	log.Info("got discover movies", "count", len(m.Results))

	res := make(types.Content, 0, len(m.Results))

	for _, v := range m.Results {
		mr := converter.MovieByGenreResult(v)
		ci, err := mr.Convert(a.cfg.Urls.TMDbImageUrl)
		if err != nil {
			log.Warn("movie result convert error", "id", v.ID, "error", err.Error())
			continue
		}

		res = append(res, ci)
	}

	return res, nil
}
