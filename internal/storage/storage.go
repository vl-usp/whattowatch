package storage

import (
	"context"
	"fmt"
	"log/slog"
	"whattowatch/internal/config"
	"whattowatch/internal/storage/postgresql"
	"whattowatch/internal/types"

	"github.com/gofrs/uuid"
)

type ContentStorer interface {
	GetContent(ctx context.Context, id uuid.UUID) (types.Content, error)
	GetContentTMDbIDs(ctx context.Context) ([]uuid.UUID, error)
	GetContentByTitles(ctx context.Context, titles []string) (types.Contents, error)
	InsertContent(ctx context.Context, content types.Content) error
	InsertContents(ctx context.Context, contents types.Contents) error
	UpdateContent(ctx context.Context, content types.Content) error
}

type GenreStorer interface {
	GetContentGenres(ctx context.Context, filmContentID uuid.UUID) (types.Genres, error)
	GetGenresByIDs(ctx context.Context, ids []int) (types.Genres, error)
	InsertGenre(ctx context.Context, genre types.Genre) error
	InsertGenres(ctx context.Context, genres types.Genres) error
	InsertContentGenres(ctx context.Context, filmContentID uuid.UUID, tmdbGenreIDs []int32) error
}

type UserStorer interface {
	GetUser(ctx context.Context, id int) (types.User, error)
	InsertUser(ctx context.Context, user types.User) error
}

type FavoriteStorer interface {
	GetUserFavorites(ctx context.Context, userID int) (types.Contents, error)
	GetUserFavoritesIDs(ctx context.Context, userID int) ([]uuid.UUID, error)
	GetUserFavoriteIDByTitle(ctx context.Context, userID int, title string) (uuid.UUID, error)
	GetUserFavoritesByType(ctx context.Context, userID int) (types.ContentsByTypes, error)
	InsertUserFavorites(ctx context.Context, userID int, filmContentIDs []uuid.UUID) error
	DeleteUserFavorites(ctx context.Context, userID int, filmContentIDs []uuid.UUID) error
}

type ViewedStorer interface {
	InsertUserViewed(ctx context.Context, userID int, filmContentID uuid.UUID) error
	InsertUserVieweds(ctx context.Context, userID int, filmContentIDs []uuid.UUID) error
	GetUserViewed(ctx context.Context, userID int) (types.UserVieweds, error)
}

type Storer interface {
	ContentStorer
	GenreStorer
	UserStorer
	FavoriteStorer
}

func New(cfg *config.Config, log *slog.Logger) (Storer, error) {
	switch cfg.StorageType {
	case "postgresql":
		return postgresql.New(cfg, log)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", cfg.StorageType)
	}
}
