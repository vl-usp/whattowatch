package types

import "github.com/gofrs/uuid"

type UserViewed struct {
	UserID     int
	ContentIDs []uuid.UUID
}
