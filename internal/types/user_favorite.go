package types

import "github.com/gofrs/uuid"

type UserFavorite struct {
	ID         int
	UserID     int
	FavoriteID uuid.UUID
}

type UserFavorites []UserFavorite

func (t UserFavorites) FavoriteIDs() []uuid.UUID {
	res := make([]uuid.UUID, 0, len(t))
	for _, movie := range t {
		res = append(res, movie.FavoriteID)
	}
	return res
}
