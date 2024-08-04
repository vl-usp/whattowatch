package types

import "github.com/google/uuid"

type UserFavorite struct {
	ID              int
	UserID          int
	FilmContentType int
	FavoriteID      uuid.UUID
}

type UserFavorites []UserFavorite
