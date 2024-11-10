package types

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Genre struct {
	ID   int64
	Name string
}

func (g Genre) String() string {
	return g.Name
}

type Genres []Genre

func (g Genres) String() string {
	names := make([]string, 0, len(g))
	for _, genre := range g {
		names = append(names, genre.String())
	}
	return strings.Join(names, ", ")
}

func (g Genres) GetIDs() []int64 {
	ids := make([]int64, 0, len(g))
	for _, genre := range g {
		ids = append(ids, genre.ID)
	}
	return ids
}

func (genres Genres) GetInfo(contentType ContentType) string {
	builder := strings.Builder{}

	var prefix, title string

	switch contentType {
	case Movie:
		prefix = "/gf"
		title = "*Фильмы. Выберите жанр:*"
	case TV:
		prefix = "/gt"
		title = "*Сериалы. Выберите жанр:*"
	}

	caser := cases.Title(language.Russian)

	builder.WriteString(fmt.Sprintf("%s\n", title))
	for _, g := range genres {
		builder.WriteString(fmt.Sprintf("%s (%s%d)\n", caser.String(g.Name), prefix, g.ID))
	}
	return builder.String()
}
