package storage

import (
	"context"
	"fmt"
	"log/slog"
	"whattowatch/internal/config"
	"whattowatch/internal/storage/postgresql"
	"whattowatch/internal/types"

	"github.com/google/uuid"
)

type FilmContentStorer interface {
	GetFilmContent(ctx context.Context, id uuid.UUID) (types.FilmContent, error)
	GetFilmContentTMDbIDs(ctx context.Context) ([]uuid.UUID, error)
	GetFilmContentByTitles(ctx context.Context, titles []string) (types.FilmContents, error)
	UpdateFilmContent(ctx context.Context, content types.FilmContent) error
	InsertFilmContent(ctx context.Context, content types.FilmContent) error
	InsertFilmContents(ctx context.Context, contents types.FilmContents) error
}

type FilmGenreStorer interface {
	InsertFilmGenre(ctx context.Context, genre types.FilmGenre) error
	InsertFilmGenres(ctx context.Context, genres []types.FilmGenre) error
	InsertFilmContentGenres(ctx context.Context, filmContentID uuid.UUID, tmdbGenreIDs []int32) error
}

type UserStorer interface {
	GetUser(ctx context.Context, id int) (types.User, error)
	InsertUser(ctx context.Context, user types.User) error
}

type FavoriteStorer interface {
	// Favorites
	InsertUserFavorites(ctx context.Context, userID int, filmContentIds []uuid.UUID) error
	GetUserFavorites(ctx context.Context, userID int) (types.FilmContents, error)
	GetFilmContentByTitles(ctx context.Context, titles []string) (types.FilmContents, error)
}

type Storer interface {
	FilmContentStorer
	FilmGenreStorer
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
