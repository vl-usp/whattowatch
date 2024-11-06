package storage

import (
	"context"
	"fmt"
	"log/slog"
	"whattowatch/internal/config"
	"whattowatch/internal/storage/postgresql"
	"whattowatch/internal/types"
)

type ContentStorer interface {
	GetContentItem(ctx context.Context, id int64) (types.ContentItem, error)
	InsertContent(ctx context.Context, contents types.Content) error
	GetContentStatus(ctx context.Context, userID int64, contentID int64) (types.ContentStatus, error)

	FavoriteContentStorer
	ViewedContentStorer
}

type FavoriteContentStorer interface {
	AddContentItemToFavorite(ctx context.Context, userID int64, contentID int64) error
	RemoveContentItemFromFavorite(ctx context.Context, userID int64, contentID int64) error

	GetFavoriteContent(ctx context.Context, userID int64, contentType types.ContentType) (types.Content, error)
}

type ViewedContentStorer interface {
	AddContentItemToViewed(ctx context.Context, userID int64, contentID int64) error
	RemoveContentItemFromViewed(ctx context.Context, userID int64, contentID int64) error

	GetViewedContent(ctx context.Context, userID int64, contentType types.ContentType) (types.Content, error)
}

type GenreStorer interface {
	GetGenres(ctx context.Context, contentID int64) (types.Genres, error)
	GetGenresByIDs(ctx context.Context, ids []int64) (types.Genres, error)
	InsertGenres(ctx context.Context, genres types.Genres) error
	InsertContentGenres(ctx context.Context, contentID int64, tmdbGenreIDs []int64) error
}

type UserStorer interface {
	GetUser(ctx context.Context, id int) (types.User, error)
	InsertUser(ctx context.Context, user types.User) error
}

type Storer interface {
	ContentStorer
	GenreStorer
	UserStorer
}

func New(cfg *config.Config, log *slog.Logger) (Storer, error) {
	switch cfg.StorageType {
	case "postgresql":
		return postgresql.New(cfg, log)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", cfg.StorageType)
	}
}
