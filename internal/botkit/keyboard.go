package botkit

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/reply"
)

var replyKeyboard *reply.ReplyKeyboard

func initReplyKeyboard(b *bot.Bot) {
	replyKeyboard = reply.New(
		b,
		reply.WithPrefix("reply_keyboard"),
		reply.IsSelective(),
	).
		Button("Фильмы", b, bot.MatchTypeExact, onMoviesKeyboard).
		Button("Сериалы", b, bot.MatchTypeExact, onTVsKeyboard)
}

func onMainKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	replyKeyboard = reply.New(
		b,
		reply.WithPrefix("reply_keyboard_main"),
		reply.IsSelective(),
	).
		Button("Фильмы", b, bot.MatchTypeExact, onMoviesKeyboard).
		Button("Сериалы", b, bot.MatchTypeExact, onTVsKeyboard)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Выберите тип контента, который хотите посмотреть:",
		ReplyMarkup: replyKeyboard,
	})
}

func onMoviesKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	replyKeyboard = reply.New(
		b,
		reply.WithPrefix("reply_keyboard_movies"),
		reply.IsSelective(),
	).
		Button("Рекомендации", b, bot.MatchTypeExact, onMoviesRecomendations).
		Button("Популярные", b, bot.MatchTypeExact, onMoviesPopular).
		Button("Лучшие", b, bot.MatchTypeExact, onMoviesTop).
		Button("Просмотренные", b, bot.MatchTypeExact, onMoviesViewed).
		Row().
		Button("Назад", b, bot.MatchTypePrefix, onMainKeyboard)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Фильмы. Выберите раздел раздел:",
		ReplyMarkup: replyKeyboard,
	})
}

func onMoviesRecomendations(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO: вывод рекомендаций
}

func onMoviesPopular(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO: вывод лучших
}

func onMoviesTop(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO: вывод лучших
}

func onMoviesViewed(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO: вывод просмотренных
}

func onTVsKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	replyKeyboard = reply.New(
		b,
		reply.WithPrefix("reply_keyboard_tvs"),
		reply.IsSelective(),
	).
		Button("Рекомендации", b, bot.MatchTypeExact, onTVsRecomendations).
		Button("Популярные", b, bot.MatchTypeExact, onTVsPopular).
		Button("Лучшие", b, bot.MatchTypeExact, onTVsTop).
		Button("Просмотренные", b, bot.MatchTypeExact, onTVsViewed).
		Row().
		Button("Назад", b, bot.MatchTypePrefix, onMainKeyboard)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Сериалы. Выберите раздел:",
		ReplyMarkup: replyKeyboard,
	})
}

func onTVsRecomendations(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO: вывод рекомендаций
}

func onTVsPopular(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO: вывод лучших
}

func onTVsTop(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO: вывод лучших
}

func onTVsViewed(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO: вывод просмотренных
}

func (t *TGBot) handlerReplyKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	t.log.With("fn", "handlerReplyKeyboard", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID).Debug("handler func start log")

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Выберите тип контента, который хотите посмотреть:",
		ReplyMarkup: replyKeyboard,
	})
}

func onReplyKeyboardSelect(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "You selected: " + string(update.Message.Text),
	})
}
