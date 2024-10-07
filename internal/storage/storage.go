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
	GetGenres(ctx context.Context, contentID uuid.UUID) (types.Genres, error)
	GetGenresByIDs(ctx context.Context, ids []int) (types.Genres, error)
	InsertGenre(ctx context.Context, genre types.Genre) error
	InsertGenres(ctx context.Context, genres types.Genres) error
	InsertContentGenres(ctx context.Context, contentID uuid.UUID, tmdbGenreIDs []int32) error
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
