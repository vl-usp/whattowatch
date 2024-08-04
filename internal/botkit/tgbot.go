package botkit

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"whattowatch/internal/config"
	"whattowatch/internal/storage"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type HandlerFunc func(ctx context.Context, b *bot.Bot, update *models.Update)

type TGBot struct {
	storer   storage.Storer
	bot      *bot.Bot
	handlers map[string]HandlerFunc

	log *slog.Logger
	cfg *config.Config
}

func NewTGBot(cfg *config.Config, log *slog.Logger, storer storage.Storer) (*TGBot, error) {
	tgbot := &TGBot{
		storer: storer,

		log: log,
		cfg: cfg,
	}
	opts := []bot.Option{
		bot.WithDebug(),
		// bot.With
		// bot.WithMiddlewares(tgbot.getUserMiddleware),
		// bot.WithDefaultHandler(tgbot.defaultHandler),
	}
	b, err := bot.New(cfg.Tokens.TGBot, opts...)
	if err != nil {
		return nil, err
	}
	tgbot.bot = b

	tgbot.addHandlers()
	return tgbot, nil
}

func (t *TGBot) Start() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	t.bot.Start(ctx)
}

func (t *TGBot) addHandlers() {
	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, t.helpHandler)
	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, t.registerHandler)
	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/menu", bot.MatchTypeExact, t.menuHandler)
	// t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/add-movies", bot.MatchTypePrefix, t.addFavoriteMoviesHandler)
	// t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/add-series", bot.MatchTypePrefix, t.addFavoriteTVsHandler)
	// t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/recommend", bot.MatchTypeExact, t.recommendHandler)
	// t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/search-movie", bot.MatchTypePrefix, t.searchMovieHandler)
	// t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/search-series", bot.MatchTypePrefix, t.searchTVsHandler)
	// t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/show-favorites", bot.MatchTypePrefix, t.showFavoritesHandler)
}
