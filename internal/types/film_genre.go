package types

import "github.com/google/uuid"

type FilmGenre struct {
	ID     uuid.UUID
	TMDbID int
	Name   string
	Slug   string
}
