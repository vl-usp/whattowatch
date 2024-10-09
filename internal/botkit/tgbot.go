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

	// TODO keyboard map[user_id]keyboard
	// TODO pages map[user_id]struct{}
}

func NewTGBot(cfg *config.Config, log *slog.Logger, storer storage.Storer) (*TGBot, error) {
	api, err := api.New(cfg, storer, log)
	if err != nil {
		return nil, err
	}

	tgbot := &TGBot{
		storer: storer,
		api:    api,

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

	tgbot.useKeyboard()
	tgbot.useHandlers()

	return tgbot, nil
}

func (t *TGBot) Start() {
	log := t.log.With("fn", "Start")
	log.Info("starting bot")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	t.bot.Start(ctx)
}

func (t *TGBot) useHandlers() {
	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypePrefix, t.helpHandler)
	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypePrefix, t.registerHandler)
	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/menu", bot.MatchTypePrefix, t.handlerReplyKeyboard)

	// t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/add_favorite", bot.MatchTypePrefix, t.addFavoriteHandler)
	// t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/remove_favorite", bot.MatchTypePrefix, t.removeFavoriteHandler)

	// t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/add_viewed", bot.MatchTypePrefix, t.addViewedHandler)
	// t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/remove_viewed", bot.MatchTypePrefix, t.removeViewedHandler)

	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/m", bot.MatchTypePrefix, t.searchMovieHandler)
	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/t", bot.MatchTypePrefix, t.searchTVHandler)
}

func (t *TGBot) useKeyboard() {
	t.initReplyKeyboard(t.bot)
}
