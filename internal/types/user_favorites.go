package types

import "github.com/gofrs/uuid"

type UserFavorites struct {
	UserID      int
	FavoriteIDs []uuid.UUID
}
