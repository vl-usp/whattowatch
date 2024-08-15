package types

import (
	"strings"

	"github.com/gofrs/uuid"
)

type Genre struct {
	ID            uuid.UUID
	TMDbID        int
	Name          string
	Slug          string
	FormattedName string
}

type Genres []Genre

func (fg Genres) String() string {
	names := make([]string, 0, len(fg))
	for _, genre := range fg {
		names = append(names, genre.FormattedName)
	}
	return strings.Join(names, ", ")
}
