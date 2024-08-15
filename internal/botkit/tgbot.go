package botkit

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"whattowatch/internal/api"
	"whattowatch/internal/config"
	"whattowatch/internal/storage"

	"github.com/go-telegram/bot"
)

type TGBot struct {
	storer storage.Storer
	bot    *bot.Bot
	api    *api.TMDbApi

	log *slog.Logger
	cfg *config.Config
}

func NewTGBot(cfg *config.Config, log *slog.Logger, storer storage.Storer) (*TGBot, error) {
	tgbot := &TGBot{
		storer: storer,
		api:    api.New(cfg.Tokens.TMDb, storer, log),

		log: log.With("pkg", "botkit"),
		cfg: cfg,
	}

	opts := []bot.Option{
		// bot.WithDebug(),
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
	log := t.log.With("fn", "Start")
	log.Info("starting bot")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	t.bot.Start(ctx)
}

func (t *TGBot) addHandlers() {
	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypePrefix, t.helpHandler)
	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypePrefix, t.registerHandler)
	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/menu", bot.MatchTypePrefix, t.menuHandler)
	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/add", bot.MatchTypePrefix, t.addFavoriteHandler)
	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/remove", bot.MatchTypePrefix, t.removeFavoriteHandler)
}
