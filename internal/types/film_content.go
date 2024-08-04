package types

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type FilmContent struct {
	ID                uuid.UUID
	TMDbID            int
	FilmContentTypeId int
	Title             string
	Overview          string
	Popularity        float32
	PosterPath        string
	ReleaseDate       sql.NullTime
	VoteAverage       float32
	VoteCount         uint32
}

func GetReleaseDate(in string) (sql.NullTime, error) {
	relesaseDate, err := time.Parse("2006-01-02", in)
	if err != nil {
		return sql.NullTime{}, fmt.Errorf("parse release date from %s error: %s", in, err.Error())
	}
	return sql.NullTime{Time: relesaseDate, Valid: true}, nil
}

type FilmContents []FilmContent

func (t FilmContents) IDs() []uuid.UUID {
	res := make([]uuid.UUID, 0, len(t))
	for _, movie := range t {
		res = append(res, movie.ID)
	}
	return res
}
