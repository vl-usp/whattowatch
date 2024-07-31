package types

import (
	"fmt"
	"time"
)

type Movies []*Movie

type Movie struct {
	ID          int
	SourceID    int
	Title       string
	Description string
	Runtime     string
	ReleaseDate string
	Genres      []Genre
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   time.Time
}

func (m *Movie) String() string {
	return fmt.Sprintf(
		"ID: %d\nSourceID: %d\nTitle: %s\nDescription: %s\nRuntime: %s\nReleaseDate: %s\n\n",
		m.ID,
		m.SourceID,
		m.Title,
		m.Description,
		m.Runtime,
		m.ReleaseDate,
	)
}
