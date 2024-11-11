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

func (a *TMDbApi) GetTV(ctx context.Context, id int) (types.ContentItem, error) {
	log := a.log.With("fn", "GetTV", "id", id)

	opts := a.getOpts()
	opts["append_to_response"] = "videos"

	tv, err := a.client.GetTVDetails(id, opts)
	if err != nil {
		return types.ContentItem{}, err
	}
	log.Debug("got tv details", "id", tv.ID, "title", tv.Name)

	rd, err := time.Parse("2006-01-02", tv.FirstAirDate)
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

func (a *TMDbApi) GetTVPopular(ctx context.Context, page int) (types.Content, error) {
	log := a.log.With("fn", "GetTVPopular", "page", page)

	opts := a.getOpts()
	opts["page"] = fmt.Sprintf("%d", page)

	m, err := a.client.GetTVPopular(opts)
	if err != nil {
		return nil, err
	}
	log.Info("got tv popular", "count", len(m.Results))

	res := make(types.Content, 0, len(m.Results))
	for _, v := range m.Results {
		tr := converter.TVPageResult(v)
		ci, err := tr.Convert(a.cfg.Urls.TMDbImageUrl)
		if err != nil {
			log.Warn("tv result convert error", "id", v.ID, "error", err.Error())
			continue
		}

		res = append(res, ci)
	}

	return res, nil
}

func (a *TMDbApi) GetTVTop(ctx context.Context, page int) (types.Content, error) {
	log := a.log.With("fn", "GetTVTop", "page", page)

	opts := a.getOpts()
	opts["page"] = fmt.Sprintf("%d", page)

	m, err := a.client.GetTVTopRated(opts)
	if err != nil {
		return nil, err
	}
	log.Info("got tv top rated", "count", len(m.Results))

	res := make(types.Content, 0, len(m.Results))
	for _, v := range m.Results {
		tr := converter.TVPageResult(v)
		ci, err := tr.Convert(a.cfg.Urls.TMDbImageUrl)
		if err != nil {
			log.Warn("tv result convert error", "id", v.ID, "error", err.Error())
			continue
		}

		res = append(res, ci)
	}

	return res, nil
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
					log.Error("request error", "id", id, "tv_id", job, "error", err.Error())
					tvCh <- content{
						err: err,
					}
				}

				c := make(types.Content, 0, len(res.Results))
				for _, v := range res.Results {
					tr := converter.TVRecommendationResult(v)
					ci, err := tr.Convert(a.cfg.Urls.TMDbImageUrl)
					if err != nil {
						log.Warn("tv result convert error", "id", v.ID, "error", err.Error())
						continue
					}

					c = append(c, ci)
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

					tr := converter.TVSearchResult(v)
					ci, err := tr.Convert(a.cfg.Urls.TMDbImageUrl)
					if err != nil {
						log.Warn("tv result convert error", "id", v.ID, "error", err.Error())
						continue
					}

					c = append(c, ci)
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

func (a *TMDbApi) GetTVsByGenre(ctx context.Context, genreIDs []int, page int) (types.Content, error) {
	log := a.log.With("fn", "DiscoverTV", "page", page, "genres", genreIDs)

	opts := a.getOpts()
	opts["page"] = fmt.Sprintf("%d", page)
	opts["with_genres"] = strings.Join(utils.IntSliceToStringSlice(genreIDs), ",")

	m, err := a.client.GetDiscoverTV(opts)
	if err != nil {
		return nil, err
	}
	log.Info("got discover tv", "count", len(m.Results))

	c := make(types.Content, 0, len(m.Results))

	for _, v := range m.Results {
		tr := converter.TVByGenreResult(v)
		ci, err := tr.Convert(a.cfg.Urls.TMDbImageUrl)
		if err != nil {
			log.Warn("tv result convert error", "id", v.ID, "error", err.Error())
			continue
		}

		c = append(c, ci)
	}

	return c, nil
}
