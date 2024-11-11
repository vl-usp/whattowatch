package converter

import (
	"errors"
	"time"
	"whattowatch/internal/types"
)

type MoviePageResult struct {
	PosterPath  string
	Adult       bool
	Overview    string
	ReleaseDate string
	Genres      []struct {
		ID   int64
		Name string
	}
	ID               int64
	OriginalTitle    string
	OriginalLanguage string
	Title            string
	BackdropPath     string
	Popularity       float32
	VoteCount        int64
	Video            bool
	VoteAverage      float32
}

func (mr MoviePageResult) Convert(imageUrl string) (types.ContentItem, error) {
	rd, err := time.Parse("2006-01-02", mr.ReleaseDate)
	if err != nil {
		return types.ContentItem{}, err
	}

	title := mr.Title
	if mr.Title == "" {
		title = mr.OriginalTitle
	}

	poster := imageUrl + mr.PosterPath
	if mr.PosterPath == "" {
		poster = emptyImageUrl
	}

	backdrop := imageUrl + mr.BackdropPath
	if mr.BackdropPath == "" {
		backdrop = emptyImageUrl
	}

	return types.ContentItem{
		ID:           mr.ID,
		ContentType:  types.Movie,
		Title:        title,
		Overview:     mr.Overview,
		Popularity:   mr.Popularity,
		PosterPath:   poster,
		BackdropPath: backdrop,
		ReleaseDate:  rd,
		VoteAverage:  mr.VoteAverage,
		VoteCount:    mr.VoteCount,
	}, nil
}

type MovieRecommendationResult struct {
	PosterPath       string
	Adult            bool
	Overview         string
	ReleaseDate      string
	GenreIDs         []int64
	ID               int64
	OriginalTitle    string
	OriginalLanguage string
	Title            string
	BackdropPath     string
	Popularity       float32
	VoteCount        int64
	Video            bool
	VoteAverage      float32
}

func (mr MovieRecommendationResult) Convert(imageUrl string) (types.ContentItem, error) {
	if mr.Overview == "" {
		return types.ContentItem{}, errors.New("empty overview")
	}

	rd, err := time.Parse("2006-01-02", mr.ReleaseDate)
	if err != nil {
		return types.ContentItem{}, err
	}

	filterTime, err := time.Parse("2006-01-02", recomendationDateFrom)
	if err != nil {
		return types.ContentItem{}, err
	}

	if rd.Before(filterTime) {
		return types.ContentItem{}, errors.New("old movie. skip")
	}

	title := mr.Title
	if mr.Title == "" {
		title = mr.OriginalTitle
	}

	poster := imageUrl + mr.PosterPath
	if mr.PosterPath == "" {
		poster = emptyImageUrl
	}

	backdrop := imageUrl + mr.BackdropPath
	if mr.BackdropPath == "" {
		backdrop = emptyImageUrl
	}

	return types.ContentItem{
		ID:           mr.ID,
		ContentType:  types.Movie,
		Title:        title,
		Overview:     mr.Overview,
		Popularity:   mr.Popularity,
		PosterPath:   poster,
		BackdropPath: backdrop,
		ReleaseDate:  rd,
		VoteAverage:  mr.VoteAverage,
		VoteCount:    mr.VoteCount,
	}, nil
}

type MovieSearchResult struct {
	VoteCount        int64
	ID               int64
	Video            bool
	VoteAverage      float32
	Title            string
	Popularity       float32
	PosterPath       string
	OriginalLanguage string
	OriginalTitle    string
	GenreIDs         []int64
	BackdropPath     string
	Adult            bool
	Overview         string
	ReleaseDate      string
}

func (mr MovieSearchResult) Convert(imageUrl string) (types.ContentItem, error) {
	rd, err := time.Parse("2006-01-02", mr.ReleaseDate)
	if err != nil {
		return types.ContentItem{}, err
	}

	title := mr.Title
	if mr.Title == "" {
		title = mr.OriginalTitle
	}

	poster := imageUrl + mr.PosterPath
	if mr.PosterPath == "" {
		poster = emptyImageUrl
	}

	backdrop := imageUrl + mr.BackdropPath
	if mr.BackdropPath == "" {
		backdrop = emptyImageUrl
	}

	return types.ContentItem{
		ID:           mr.ID,
		ContentType:  types.Movie,
		Title:        title,
		Overview:     mr.Overview,
		Popularity:   mr.Popularity,
		PosterPath:   poster,
		BackdropPath: backdrop,
		ReleaseDate:  rd,
		VoteAverage:  mr.VoteAverage,
		VoteCount:    mr.VoteCount,
	}, nil
}

type MovieByGenreResult struct {
	VoteCount        int64
	ID               int64
	Video            bool
	VoteAverage      float32
	Title            string
	Popularity       float32
	PosterPath       string
	OriginalLanguage string
	OriginalTitle    string
	GenreIDs         []int64
	BackdropPath     string
	Adult            bool
	Overview         string
	ReleaseDate      string
}

func (mr MovieByGenreResult) Convert(imageUrl string) (types.ContentItem, error) {
	rd, err := time.Parse("2006-01-02", mr.ReleaseDate)
	if err != nil {
		return types.ContentItem{}, err
	}

	title := mr.Title
	if mr.Title == "" {
		title = mr.OriginalTitle
	}

	poster := imageUrl + mr.PosterPath
	if mr.PosterPath == "" {
		poster = emptyImageUrl
	}

	backdrop := imageUrl + mr.BackdropPath
	if mr.BackdropPath == "" {
		backdrop = emptyImageUrl
	}

	return types.ContentItem{
		ID:           mr.ID,
		ContentType:  types.Movie,
		Title:        title,
		Overview:     mr.Overview,
		Popularity:   mr.Popularity,
		PosterPath:   poster,
		BackdropPath: backdrop,
		ReleaseDate:  rd,
		VoteAverage:  mr.VoteAverage,
		VoteCount:    mr.VoteCount,
	}, nil
}
