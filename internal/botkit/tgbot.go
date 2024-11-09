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
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/reply"
)

type UserData struct {
	replyKeyboard *reply.ReplyKeyboard

	popularMoviesPage int
	topMoviePage      int

	popularTVsPage int
	topTVsPage     int
}

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
)

type TGBot struct {
	storer  Storer
	content ContentProvider

	bot *bot.Bot

	log *slog.Logger
	cfg *config.Config

	userData map[int64]UserData
	mu       sync.RWMutex
}

func NewTGBot(cfg *config.Config, log *slog.Logger, storer Storer, contentProvider ContentProvider) (*TGBot, error) {

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

func (t *TGBot) userDataMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		log := t.log.With("fn", "userDataMiddleware")

		var id int64
		if update.CallbackQuery != nil {
			id = update.CallbackQuery.From.ID
		} else {
			id = update.Message.From.ID
		}

		t.mu.RLock()
		entry, ok := t.userData[id]
		t.mu.RUnlock()

		if !ok {
			log.Debug("init user data", "userID", id)

			ud := UserData{
				replyKeyboard: t.getMainKeyboard(),

				popularMoviesPage: 1,
				topMoviePage:      1,

				popularTVsPage: 1,
				topTVsPage:     1,
			}

			t.mu.Lock()
			t.userData[id] = ud
			t.mu.Unlock()

			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      id,
				Text:        "Выберите тип контента, который хотите посмотреть",
				ReplyMarkup: entry.replyKeyboard,
			})
		}

		next(ctx, b, update)
	}
}
