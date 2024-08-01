package tgbot

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"whattowatch/internal/config"

	"github.com/go-telegram/bot"
)

type TGBot struct {
	log *slog.Logger
	cfg *config.Config
	bot *bot.Bot
}

func New(cfg *config.Config, log *slog.Logger) (*TGBot, error) {
	tgbot := &TGBot{
		log: log,
		cfg: cfg,
	}
	opts := []bot.Option{
		bot.WithMiddlewares(tgbot.showMessageWithUserID, tgbot.showMessageWithUserName),
		bot.WithDefaultHandler(tgbot.defaultHandler),
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
	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, t.startHandler)
	// t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, t.helpHandler)
}
