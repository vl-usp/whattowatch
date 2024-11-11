package converter

import (
	"errors"
	"time"
	"whattowatch/internal/types"
)

type TVPageResult struct {
	OriginalName     string
	GenreIDs         []int64
	Name             string
	Popularity       float32
	OriginCountry    []string
	VoteCount        int64
	FirstAirDate     string
	BackdropPath     string
	OriginalLanguage string
	ID               int64
	VoteAverage      float32
	Overview         string
	PosterPath       string
}

func (tr *TVPageResult) Convert(imageUrl string) (types.ContentItem, error) {
	rd, err := time.Parse("2006-01-02", tr.FirstAirDate)
	if err != nil {
		return types.ContentItem{}, err
	}

	title := tr.Name
	if tr.Name == "" {
		title = tr.OriginalName
	}

	poster := imageUrl + tr.PosterPath
	if tr.PosterPath == "" {
		poster = emptyImageUrl
	}

	backdrop := imageUrl + tr.BackdropPath
	if tr.BackdropPath == "" {
		backdrop = emptyImageUrl
	}

	return types.ContentItem{
		ID:           tr.ID,
		ContentType:  types.TV,
		Title:        title,
		Overview:     tr.Overview,
		Popularity:   tr.Popularity,
		PosterPath:   poster,
		BackdropPath: backdrop,
		ReleaseDate:  rd,
		VoteAverage:  tr.VoteAverage,
		VoteCount:    tr.VoteCount,
	}, nil
}

type TVRecommendationResult struct {
	PosterPath       string
	Popularity       float32
	ID               int64
	BackdropPath     string
	VoteAverage      float32
	Overview         string
	FirstAirDate     string
	OriginCountry    []string
	GenreIDs         []int64
	OriginalLanguage string
	VoteCount        int64
	Name             string
	OriginalName     string
}

func (tr *TVRecommendationResult) Convert(imageUrl string) (types.ContentItem, error) {
	if tr.Overview == "" {
		return types.ContentItem{}, errors.New("empty overview")
	}

	rd, err := time.Parse("2006-01-02", tr.FirstAirDate)
	if err != nil {
		return types.ContentItem{}, err
	}

	filterTime, err := time.Parse("2006-01-02", recomendationDateFrom)
	if err != nil {
		return types.ContentItem{}, err
	}

	if rd.Before(filterTime) {
		return types.ContentItem{}, errors.New("old tv series. skip")
	}

	title := tr.Name
	if tr.Name == "" {
		title = tr.OriginalName
	}

	poster := imageUrl + tr.PosterPath
	if tr.PosterPath == "" {
		poster = emptyImageUrl
	}

	backdrop := imageUrl + tr.BackdropPath
	if tr.BackdropPath == "" {
		backdrop = emptyImageUrl
	}

	return types.ContentItem{
		ID:           tr.ID,
		ContentType:  types.TV,
		Title:        title,
		Overview:     tr.Overview,
		Popularity:   tr.Popularity,
		PosterPath:   poster,
		BackdropPath: backdrop,
		ReleaseDate:  rd,
		VoteAverage:  tr.VoteAverage,
		VoteCount:    tr.VoteCount,
	}, nil
}

type TVSearchResult struct {
	OriginalName     string
	ID               int64
	Name             string
	VoteCount        int64
	VoteAverage      float32
	PosterPath       string
	FirstAirDate     string
	Popularity       float32
	GenreIDs         []int64
	OriginalLanguage string
	BackdropPath     string
	Overview         string
	OriginCountry    []string
}

func (tr TVSearchResult) Convert(imageUrl string) (types.ContentItem, error) {
	rd, err := time.Parse("2006-01-02", tr.FirstAirDate)
	if err != nil {
		return types.ContentItem{}, err
	}

	title := tr.Name
	if tr.Name == "" {
		title = tr.OriginalName
	}

	poster := imageUrl + tr.PosterPath
	if tr.PosterPath == "" {
		poster = emptyImageUrl
	}

	backdrop := imageUrl + tr.BackdropPath
	if tr.BackdropPath == "" {
		backdrop = emptyImageUrl
	}

	return types.ContentItem{
		ID:           tr.ID,
		ContentType:  types.TV,
		Title:        title,
		Overview:     tr.Overview,
		Popularity:   tr.Popularity,
		PosterPath:   poster,
		BackdropPath: backdrop,
		ReleaseDate:  rd,
		VoteAverage:  tr.VoteAverage,
		VoteCount:    tr.VoteCount,
	}, nil
}

type TVByGenreResult struct {
	OriginalName     string
	GenreIDs         []int64
	Name             string
	Popularity       float32
	OriginCountry    []string
	VoteCount        int64
	FirstAirDate     string
	BackdropPath     string
	OriginalLanguage string
	ID               int64
	VoteAverage      float32
	Overview         string
	PosterPath       string
}

func (tr TVByGenreResult) Convert(imageUrl string) (types.ContentItem, error) {
	rd, err := time.Parse("2006-01-02", tr.FirstAirDate)
	if err != nil {
		return types.ContentItem{}, err
	}

	title := tr.Name
	if tr.Name == "" {
		title = tr.OriginalName
	}

	poster := imageUrl + tr.PosterPath
	if tr.PosterPath == "" {
		poster = emptyImageUrl
	}

	backdrop := imageUrl + tr.BackdropPath
	if tr.BackdropPath == "" {
		backdrop = emptyImageUrl
	}

	return types.ContentItem{
		ID:           tr.ID,
		ContentType:  types.TV,
		Title:        title,
		Overview:     tr.Overview,
		Popularity:   tr.Popularity,
		PosterPath:   poster,
		BackdropPath: backdrop,
		ReleaseDate:  rd,
		VoteAverage:  tr.VoteAverage,
		VoteCount:    tr.VoteCount,
	}, nil
}
