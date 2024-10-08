package types

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

type Content struct {
	ID          uuid.UUID
	ContentType ContentType
	Title       string
	Overview    string
	Popularity  float32
	PosterPath  string
	ReleaseDate sql.NullTime
	VoteAverage float32
	VoteCount   int64
	Genres      Genres
	TMDbID      int64
}

func GetReleaseDate(in string) (sql.NullTime, error) {
	relesaseDate, err := time.Parse("2006-01-02", in)
	if err != nil {
		return sql.NullTime{}, fmt.Errorf("parse release date from %s error: %s", in, err.Error())
	}
	return sql.NullTime{Time: relesaseDate, Valid: true}, nil
}

type ContentSlice []Content

func (cs ContentSlice) IDs() []uuid.UUID {
	res := make([]uuid.UUID, 0, len(cs))
	for _, movie := range cs {
		res = append(res, movie.ID)
	}
	return res
}

func (cs ContentSlice) Titles() []string {
	res := make([]string, 0, len(cs))
	for _, movie := range cs {
		res = append(res, movie.Title)
	}
	return res
}

func (cs ContentSlice) ContentTitilesMap() map[int][]string {
	res := make(map[int][]string, len(cs))
	for _, c := range cs {
		contentTypeID := c.ContentType.EnumIndex()

		if _, ok := res[contentTypeID]; !ok {
			res[contentTypeID] = make([]string, 0)
		}
		res[contentTypeID] = append(res[contentTypeID], c.Title)
	}
	return res
}

func (cs ContentSlice) PrintByContentType(prefix string) string {
	builder := strings.Builder{}
	for _, c := range cs {
		switch c.ContentType {
		case Movie:
			builder.WriteString(fmt.Sprintf("%s %s\n", prefix+" фильм", c.Title))
		case TV:
			builder.WriteString(fmt.Sprintf("%s %s\n", prefix+" сериал", c.Title))
		}
	}
	return builder.String()
}

type ContentsByTypes map[ContentType]ContentSlice

type ContentsByGenres map[string]ContentSlice
