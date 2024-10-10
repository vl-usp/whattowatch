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

	// tgbot.useKeyboard()
	tgbot.useHandlers()

	return tgbot, nil
}

func (t *TGBot) Start() {
	log := t.log.With("fn", "Start")
	bot, err := t.bot.GetMe(context.Background())
	if err != nil {
		log.Error("failed to get bot info", "error", err.Error())
		return
	}
	log.Info("starting bot", "bot_id", bot.ID)
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

	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/f", bot.MatchTypePrefix, t.searchMovieHandler)
	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/t", bot.MatchTypePrefix, t.searchTVHandler)
}

// func (t *TGBot) useKeyboard() {
// 	t.initReplyKeyboard(t.bot)
// }

func (t *TGBot) userDataMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		log := t.log.With("fn", "userDataMiddleware")

		var id int64
		if update.CallbackQuery != nil {
			id = update.CallbackQuery.From.ID
		} else {
			id = update.Message.From.ID
		}

		if entry, ok := t.userData[id]; !ok {
			log.Debug("init user data", "userID", id)

			rk := reply.New(
				b,
				reply.WithPrefix("rk_main"),
				reply.IsSelective(),
			).
				Button("Фильмы", b, bot.MatchTypeExact, t.onMoviesKeyboard).
				Button("Сериалы", b, bot.MatchTypeExact, t.onTVsKeyboard)

			ud := UserData{
				replyKeyboard: rk,

				popularMoviesPage: 1,
				topMoviePage:      1,

				popularTVsPage: 1,
				topTVsPage:     1,
			}

			t.userData[id] = ud

			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      id,
				Text:        "Выберите тип контента, который хотите посмотреть",
				ReplyMarkup: entry.replyKeyboard,
			})
		}

		next(ctx, b, update)
	}
}
