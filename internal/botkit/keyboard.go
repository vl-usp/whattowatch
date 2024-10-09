package botkit

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/inline"
	"github.com/go-telegram/ui/keyboard/reply"
)

var (
	replyKeyboard *reply.ReplyKeyboard

	popularMoviesPage int = 1
	topMoviePage      int = 1

	popularTVsPage int = 1
	topTVsPage     int = 1
)

func (t *TGBot) initReplyKeyboard(b *bot.Bot) {
	replyKeyboard = reply.New(
		b,
		reply.WithPrefix("rk"),
		reply.IsSelective(),
	).
		Button("Фильмы", b, bot.MatchTypeExact, t.onMoviesKeyboard).
		Button("Сериалы", b, bot.MatchTypeExact, t.onTVsKeyboard)
}

func (t *TGBot) onMainKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	replyKeyboard = reply.New(
		b,
		reply.WithPrefix("rk_main"),
		reply.IsSelective(),
	).
		Button("Фильмы", b, bot.MatchTypeExact, t.onMoviesKeyboard).
		Button("Сериалы", b, bot.MatchTypeExact, t.onTVsKeyboard)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Выберите тип контента, который хотите посмотреть:",
		ReplyMarkup: replyKeyboard,
	})
}

func (t *TGBot) onMoviesKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	replyKeyboard = reply.New(
		b,
		reply.WithPrefix("rk_movies"),
		reply.IsSelective(),
	).
		Button("Рекомендации", b, bot.MatchTypeExact, t.onMoviesRecomendations).
		Button("Популярные", b, bot.MatchTypeExact, t.onMoviesPopular).
		Button("Лучшие", b, bot.MatchTypeExact, t.onMoviesTop).
		Button("Просмотренные", b, bot.MatchTypeExact, t.onMoviesViewed).
		Row().
		Button("Назад", b, bot.MatchTypePrefix, t.onMainKeyboard)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Фильмы. Выберите раздел раздел:",
		ReplyMarkup: replyKeyboard,
	})
}

func (t *TGBot) onMoviesRecomendations(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO: вывод рекомендаций
}

func (t *TGBot) onMoviesPopular(ctx context.Context, b *bot.Bot, update *models.Update) {
	m, err := t.api.GetMoviesPopular(ctx, popularMoviesPage)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Ошибка...",
			ReplyMarkup: replyKeyboard,
		})
	}

	kb := inline.New(b).
		Row().
		Button("Назад", []byte("prev"), t.onMoviesPopularPage).
		Button("Далее", []byte("next"), t.onMoviesPopularPage)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        m.Print(fmt.Sprintf("Популярные фильмы #%d", popularMoviesPage)),
		ReplyMarkup: kb,
	})
}

func (t *TGBot) onMoviesPopularPage(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	if string(data) == "prev" {
		popularMoviesPage--
	} else if string(data) == "next" {
		popularMoviesPage++
	}

	if popularMoviesPage <= 0 {
		popularMoviesPage = 1
	}

	t.onMoviesPopular(ctx, b, &models.Update{
		Message: mes.Message,
	})
}

func (t *TGBot) onMoviesTop(ctx context.Context, b *bot.Bot, update *models.Update) {
	m, err := t.api.GetMovieTop(ctx, topMoviePage)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Ошибка...",
			ReplyMarkup: replyKeyboard,
		})
	}

	kb := inline.New(b).
		Row().
		Button("Назад", []byte("prev"), t.onMoviesTopPage).
		Button("Далее", []byte("next"), t.onMoviesTopPage)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        m.Print(fmt.Sprintf("Лучшие фильмы #%d", topMoviePage)),
		ReplyMarkup: kb,
	})
}

func (t *TGBot) onMoviesTopPage(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	if string(data) == "prev" {
		topMoviePage--
	} else if string(data) == "next" {
		topMoviePage++
	}

	if topMoviePage <= 0 {
		topMoviePage = 1
	}

	t.onMoviesTop(ctx, b, &models.Update{
		Message: mes.Message,
	})
}

func (t *TGBot) onMoviesViewed(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO: вывод просмотренных
}

func (t *TGBot) onTVsKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	replyKeyboard = reply.New(
		b,
		reply.WithPrefix("reply_keyboard_tvs"),
		reply.IsSelective(),
	).
		Button("Рекомендации", b, bot.MatchTypeExact, t.onTVsRecomendations).
		Button("Популярные", b, bot.MatchTypeExact, t.onTVsPopular).
		Button("Лучшие", b, bot.MatchTypeExact, t.onTVsTop).
		Button("Просмотренные", b, bot.MatchTypeExact, t.onTVsViewed).
		Row().
		Button("Назад", b, bot.MatchTypePrefix, t.onMainKeyboard)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Сериалы. Выберите раздел:",
		ReplyMarkup: replyKeyboard,
	})
}

func (t *TGBot) onTVsRecomendations(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO: вывод рекомендаций
}

func (t *TGBot) onTVsPopular(ctx context.Context, b *bot.Bot, update *models.Update) {
	m, err := t.api.GetTVPopular(ctx, popularTVsPage)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        err.Error(),
			ReplyMarkup: replyKeyboard,
		})
	}

	kb := inline.New(b).
		Row().
		Button("Назад", []byte("prev"), t.onTVsPopularPage).
		Button("Далее", []byte("next"), t.onTVsPopularPage)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        m.Print(fmt.Sprintf("Популярные сериалы #%d", popularTVsPage)),
		ReplyMarkup: kb,
	})
}

func (t *TGBot) onTVsPopularPage(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	if string(data) == "prev" {
		popularTVsPage--
	} else if string(data) == "next" {
		popularTVsPage++
	}

	if popularTVsPage <= 0 {
		popularTVsPage = 1
	}

	t.onTVsPopular(ctx, b, &models.Update{
		Message: mes.Message,
	})
}

func (t *TGBot) onTVsTop(ctx context.Context, b *bot.Bot, update *models.Update) {
	m, err := t.api.GetTVTop(ctx, topTVsPage)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        err.Error(),
			ReplyMarkup: replyKeyboard,
		})
	}

	kb := inline.New(b).
		Row().
		Button("Назад", []byte("prev"), t.onTVsTopPage).
		Button("Далее", []byte("next"), t.onTVsTopPage)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        m.Print(fmt.Sprintf("Лучшие сериалы #%d", topTVsPage)),
		ReplyMarkup: kb,
	})
}

func (t *TGBot) onTVsTopPage(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	if string(data) == "prev" {
		topTVsPage--
	} else if string(data) == "next" {
		topTVsPage++
	}

	if topTVsPage <= 0 {
		topTVsPage = 1
	}

	t.onTVsTop(ctx, b, &models.Update{
		Message: mes.Message,
	})
}

func (t *TGBot) onTVsViewed(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO: вывод просмотренных
}

func (t *TGBot) handlerReplyKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Выберите тип контента, который хотите посмотреть:",
		ReplyMarkup: replyKeyboard,
	})
}
