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
	InsertContentSlice(ctx context.Context, contents types.ContentSlice) error
}

type GenreStorer interface {
	GetGenres(ctx context.Context, contentID uuid.UUID) (types.Genres, error)
	GetGenresByIDs(ctx context.Context, ids []int) (types.Genres, error)
	InsertGenres(ctx context.Context, genres types.Genres) error
	InsertContentGenres(ctx context.Context, contentID uuid.UUID, tmdbGenreIDs []int64) error
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
