package types

import (
	"strings"
)

type Genre struct {
	ID         int64
	Name       string
	PrettyName string
}

type Genres []Genre

func (g Genres) String() string {
	names := make([]string, 0, len(g))
	for _, genre := range g {
		names = append(names, genre.PrettyName)
	}
	return strings.Join(names, ", ")
}
