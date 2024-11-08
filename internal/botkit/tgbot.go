package botkit

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"whattowatch/internal/api"
	"whattowatch/internal/config"
	"whattowatch/internal/storage"

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

type TGBot struct {
	storer storage.Storer
	bot    *bot.Bot
	api    *api.TMDbApi

	log *slog.Logger
	cfg *config.Config

	userData map[int64]UserData
	mu       sync.RWMutex
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

			rk := reply.New(
				b,
				reply.WithPrefix("rk_main"),
				reply.IsSelective(),
			).
				Button("Фильмы", b, bot.MatchTypeExact, t.onMoviesKeyboard).
				Row().
				Button("Сериалы", b, bot.MatchTypeExact, t.onTVsKeyboard)

			ud := UserData{
				replyKeyboard: rk,

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
