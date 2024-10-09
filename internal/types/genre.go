package types

import (
	"strings"
)

type Genre struct {
	ID         int64
	Name       string
	PrettyName string
}

func (g Genre) String() string {
	if g.PrettyName != "" {
		return g.PrettyName
	}
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
