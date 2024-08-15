package types

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

type Content struct {
	ID            uuid.UUID
	TMDbID        int
	ContentTypeID int
	Title         string
	Overview      string
	Popularity    float32
	PosterPath    string
	ReleaseDate   sql.NullTime
	VoteAverage   float32
	VoteCount     uint32
	Genres        Genres
}

func GetReleaseDate(in string) (sql.NullTime, error) {
	relesaseDate, err := time.Parse("2006-01-02", in)
	if err != nil {
		return sql.NullTime{}, fmt.Errorf("parse release date from %s error: %s", in, err.Error())
	}
	return sql.NullTime{Time: relesaseDate, Valid: true}, nil
}

type Contents []Content

func (t Contents) IDs() []uuid.UUID {
	res := make([]uuid.UUID, 0, len(t))
	for _, movie := range t {
		res = append(res, movie.ID)
	}
	return res
}

func (t Contents) Titles() []string {
	res := make([]string, 0, len(t))
	for _, movie := range t {
		res = append(res, movie.Title)
	}
	return res
}

func (t Contents) ContentTitilesMap() map[int][]string {
	res := make(map[int][]string, len(t))
	for _, movie := range t {
		if _, ok := res[movie.ContentTypeID]; !ok {
			res[movie.ContentTypeID] = make([]string, 0)
		}
		res[movie.ContentTypeID] = append(res[movie.ContentTypeID], movie.Title)
	}
	return res
}

func (t Contents) PrintByContentType(prefix string) string {
	builder := strings.Builder{}
	for _, fc := range t {
		switch fc.ContentTypeID {
		case 1:
			builder.WriteString(fmt.Sprintf("%s %s\n", prefix+" фильм", fc.Title))
		case 2:
			builder.WriteString(fmt.Sprintf("%s %s\n", prefix+" сериал", fc.Title))
		}
	}
	return builder.String()
}

type ContentsByTypes map[ContentType]Contents

type ContentsByGenres map[string]Contents
