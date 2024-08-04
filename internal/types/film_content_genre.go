package types

import "github.com/google/uuid"

type FilmContentGenre struct {
	ID            int
	FilmContentID uuid.UUID
	GenreID       uuid.UUID
}
