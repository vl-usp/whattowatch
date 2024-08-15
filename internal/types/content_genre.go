package types

import "github.com/gofrs/uuid"

type ContentGenre struct {
	ID        int
	ContentID uuid.UUID
	GenreID   uuid.UUID
}
