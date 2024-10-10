package types

import (
	"database/sql"
	"fmt"
	"strings"
)

type Content struct {
	ID          int64
	ContentType ContentType
	Title       string
	Overview    string
	Popularity  float32
	PosterPath  string
	ReleaseDate sql.NullTime
	VoteAverage float32
	VoteCount   int64
	Genres      Genres
}

func (c Content) String() string {
	return fmt.Sprintf(
		"ID: %d\nНазвание: %s\nЖанры: %s\nДата выхода: %s\nПопулярность: %f\nРейтинг: %f\nКоличество оценок: %d\nОписание: %s\n",
		c.ID,
		c.Title,
		c.Genres.String(),
		c.ReleaseDate.Time.Format("02.01.2006"),
		c.Popularity,
		c.VoteAverage,
		c.VoteCount,
		c.Overview,
	)
}

type ContentSlice []Content

func (cs ContentSlice) IDs() []int64 {
	res := make([]int64, 0, len(cs))
	for _, c := range cs {
		res = append(res, c.ID)
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

func (cs ContentSlice) Print(title string) string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s\n", title))
	for _, c := range cs {
		switch c.ContentType {
		case Movie:
			builder.WriteString("/f")
		case TV:
			builder.WriteString("/t")
		}
		builder.WriteString(fmt.Sprintf("%d %s (год: %d; популярность: %f; оценка: %f)\n", c.ID, c.Title, c.ReleaseDate.Time.Year(), c.Popularity, c.VoteAverage))
	}
	return builder.String()
}

type ContentsByTypes map[ContentType]ContentSlice

type ContentsByGenres map[string]ContentSlice
