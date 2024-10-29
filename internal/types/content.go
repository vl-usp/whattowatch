package types

import (
	"database/sql"
	"fmt"
	"strings"
)

type ContentItem struct {
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

func (c ContentItem) String() string {
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

type Content []ContentItem

func (content Content) Print(title string) string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s\n", title))
	for _, c := range content {
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
