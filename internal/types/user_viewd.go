package types

import "github.com/gofrs/uuid"

type UserViewed struct {
	ID        int
	UserID    int
	ContentID uuid.UUID
}

type UserVieweds []UserViewed
