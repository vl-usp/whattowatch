package types

import (
	"strings"

	"github.com/gofrs/uuid"
)

type Genre struct {
	ID         uuid.UUID
	Name       string
	PrettyName string
	TMDbID     int64
}

type Genres []Genre

func (g Genres) String() string {
	names := make([]string, 0, len(g))
	for _, genre := range g {
		names = append(names, genre.PrettyName)
	}
	return strings.Join(names, ", ")
}
