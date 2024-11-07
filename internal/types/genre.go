package types

import (
	"strings"
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
