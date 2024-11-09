package botkit

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"whattowatch/internal/config"
	"whattowatch/internal/types"

	"github.com/go-telegram/bot"
)

type (
	MovieProvider interface {
		GetMovie(ctx context.Context, id int) (types.ContentItem, error)
		GetMovies(ctx context.Context, ids []int64) (types.Content, error)
		GetMoviePopular(ctx context.Context, page int) (types.Content, error)
		GetMovieTop(ctx context.Context, page int) (types.Content, error)
		GetMovieRecommendations(ctx context.Context, ids []int64) (types.Content, error)
	}

	TVProvider interface {
		GetTV(ctx context.Context, id int) (types.ContentItem, error)
		GetTVs(ctx context.Context, ids []int64) (types.Content, error)
		GetTVPopular(ctx context.Context, page int) (types.Content, error)
		GetTVTop(ctx context.Context, page int) (types.Content, error)
		GetTVRecommendations(ctx context.Context, ids []int64) (types.Content, error)
	}

	ContentProvider interface {
		MovieProvider
		TVProvider

		SearchByTitles(ctx context.Context, titles []string) (types.Content, error)
	}

	UserStorer interface {
		InsertUser(ctx context.Context, user types.User) error
	}

	FavoriteStorer interface {
		GetFavoriteContentIDs(ctx context.Context, userID int64, contentType types.ContentType) ([]int64, error)
		AddContentItemToFavorite(ctx context.Context, userID int64, item types.ContentItem) error
		RemoveContentItemFromFavorite(ctx context.Context, userID int64, item types.ContentItem) error
	}

	ViewedStorer interface {
		GetViewedContentIDs(ctx context.Context, userID int64, contentType types.ContentType) ([]int64, error)
		AddContentItemToViewed(ctx context.Context, userID int64, item types.ContentItem) error
		RemoveContentItemFromViewed(ctx context.Context, userID int64, item types.ContentItem) error
	}

	Storer interface {
		UserStorer

		FavoriteStorer
		ViewedStorer

		GetContentStatus(ctx context.Context, userID int64, item types.ContentItem) (types.ContentStatus, error)
	}

	TGBot struct {
		storer  Storer
		content ContentProvider

		bot *bot.Bot

		log *slog.Logger
		cfg *config.Config

		userData map[int64]UserData
		mu       sync.RWMutex
	}
)

func New(cfg *config.Config, log *slog.Logger, storer Storer, contentProvider ContentProvider) (*TGBot, error) {
	tgbot := &TGBot{
		storer:  storer,
		content: contentProvider,

		log: log.With("pkg", "botkit"),
		cfg: cfg,

		userData: make(map[int64]UserData),
	}

	opts := []bot.Option{
		// bot.WithDebug(),
		bot.WithMiddlewares(tgbot.userDataMiddleware),
	}
	b, err := bot.New(cfg.Tokens.TGBot, opts...)
	if err != nil {
		return nil, err
	}
	tgbot.bot = b

	tgbot.useHandlers()

	return tgbot, nil
}

func (t *TGBot) Start() {
	log := t.log.With("fn", "Start")
	ctx := context.Background()
	bot, err := t.bot.GetMe(ctx)
	if err != nil {
		log.Error("failed to get bot info", "error", err.Error())
		return
	}
	log.Info("starting bot", "bot_id", bot.ID)
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	t.bot.Start(ctx)
}

func (t *TGBot) useHandlers() {
	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, t.helpHandler)
	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, t.registerHandler)
	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/menu", bot.MatchTypeExact, t.handlerReplyKeyboard)
	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/search", bot.MatchTypePrefix, t.searchByTitleHandler)

	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/f", bot.MatchTypePrefix, t.searchByIDHandler)
	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/t", bot.MatchTypePrefix, t.searchByIDHandler)
}

func (t *TGBot) sendErrorMessage(ctx context.Context, chatID int64) {
	t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Произошла ошибка. Попробуйте ещё раз позднее...",
	})
}
