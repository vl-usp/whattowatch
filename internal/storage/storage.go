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
	InsertContent(ctx context.Context, contents types.Content) error
	GetContentStatus(ctx context.Context, userID int64, contentID int64) (types.ContentStatus, error)
}

type FavoriteContentStorer interface {
	AddContentItemToFavorite(ctx context.Context, userID int64, contentID int64) error
	RemoveContentItemFromFavorite(ctx context.Context, userID int64, contentID int64) error

	GetFavoriteContent(ctx context.Context, userID int64) (types.Content, error)
}

type ViewedContentStorer interface {
	AddContentItemToViewed(ctx context.Context, userID int64, contentID int64) error
	RemoveContentItemFromViewed(ctx context.Context, userID int64, contentID int64) error

	GetViewedContent(ctx context.Context, userID int64) ([]int64, error)
}

type UserStorer interface {
	GetUser(ctx context.Context, id int) (types.User, error)
	InsertUser(ctx context.Context, user types.User) error
}

type Storer interface {
	ContentStorer
	UserStorer
	FavoriteContentStorer
	ViewedContentStorer
}

func New(cfg *config.Config, log *slog.Logger) (Storer, error) {
	switch cfg.StorageType {
	case "postgresql":
		return postgresql.New(cfg, log)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", cfg.StorageType)
	}
}
