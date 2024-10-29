package botkit

import (
	"context"
	"fmt"
	"whattowatch/internal/utils"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/inline"
	"github.com/go-telegram/ui/keyboard/reply"
)

// TODO: сделать основную функцию, а эту использовать как обвязку + причесать логи
func (t *TGBot) onMainKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	if entry, ok := t.userData[update.Message.From.ID]; ok {
		entry.replyKeyboard = reply.New(
			b,
			reply.WithPrefix("rk_main"),
			reply.IsSelective(),
		).
			Button("Фильмы", b, bot.MatchTypeExact, t.onMoviesKeyboard).
			Button("Сериалы", b, bot.MatchTypeExact, t.onTVsKeyboard)

		t.userData[update.Message.From.ID] = entry

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Выберите тип контента, который хотите посмотреть",
			ReplyMarkup: entry.replyKeyboard,
		})
	}
}

func (t *TGBot) onMoviesKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	if entry, ok := t.userData[update.Message.From.ID]; ok {
		entry.replyKeyboard = reply.New(
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

		t.userData[update.Message.From.ID] = entry

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Выберите тип контента, который хотите посмотреть:",
			ReplyMarkup: entry.replyKeyboard,
		})
	}
}

func (t *TGBot) onMoviesRecomendations(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO: вывод рекомендаций
	// 1) получение просмотренных
	// 2) получение избранных
	// 3) получение рекомендаций по избранным исключая просмотренные
}

func (t *TGBot) getMoviePopular(ctx context.Context, chatID int64, userData UserData) {
	m, err := t.api.GetMoviesPopular(ctx, userData.popularMoviesPage)
	if err != nil {
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        "Ошибка...",
			ReplyMarkup: userData.replyKeyboard,
		})
	}

	kb := inline.New(t.bot).
		Row().
		Button("Назад", []byte("prev"), t.onMoviesPopularPage).
		Button("Далее", []byte("next"), t.onMoviesPopularPage)

	t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        m.Print(fmt.Sprintf("Популярные фильмы #%d", userData.popularMoviesPage)),
		ReplyMarkup: kb,
	})
}

func (t *TGBot) onMoviesPopular(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "onMoviesPopular", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
	log.Debug("handler func start log")
	if entry, ok := t.userData[update.Message.From.ID]; ok {
		t.getMoviePopular(ctx, update.Message.Chat.ID, entry)
	}
}

func (t *TGBot) onMoviesPopularPage(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	log := t.log.With("fn", "onMoviesPopularPage", "user_id", mes.Message.From.ID, "chat_id", mes.Message.Chat.ID)
	log.Debug("handler func start log")
	if entry, ok := t.userData[mes.Message.Chat.ID]; ok {
		entry.popularMoviesPage = utils.HandlePage(entry.popularMoviesPage, string(data))
		t.userData[mes.Message.Chat.ID] = entry
		t.getMoviePopular(ctx, mes.Message.Chat.ID, entry)
	}
}

func (t *TGBot) getMoviesTop(ctx context.Context, chatID int64, userData UserData) {
	m, err := t.api.GetMovieTop(ctx, userData.topMoviePage)
	if err != nil {
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        "Ошибка...",
			ReplyMarkup: userData.replyKeyboard,
		})
	}

	kb := inline.New(t.bot).
		Row().
		Button("Назад", []byte("prev"), t.onMoviesTopPage).
		Button("Далее", []byte("next"), t.onMoviesTopPage)

	t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        m.Print(fmt.Sprintf("Лучшие фильмы #%d", userData.topMoviePage)),
		ReplyMarkup: kb,
	})
}

func (t *TGBot) onMoviesTop(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "onMoviesTop", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
	log.Debug("handler func start log")
	if entry, ok := t.userData[update.Message.From.ID]; ok {
		t.getMoviesTop(ctx, update.Message.Chat.ID, entry)
	}
}

func (t *TGBot) onMoviesTopPage(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	if entry, ok := t.userData[mes.Message.From.ID]; ok {
		entry.topMoviePage = utils.HandlePage(entry.topMoviePage, string(data))
		t.userData[mes.Message.From.ID] = entry
		t.getMoviesTop(ctx, mes.Message.Chat.ID, entry)
	}
}

func (t *TGBot) onMoviesViewed(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO: вывод просмотренных
}

func (t *TGBot) onTVsKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	if entry, ok := t.userData[update.Message.From.ID]; ok {
		entry.replyKeyboard = reply.New(
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

		t.userData[update.Message.From.ID] = entry

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Сериалы. Выберите раздел:",
			ReplyMarkup: entry.replyKeyboard,
		})
	}
}

func (t *TGBot) onTVsRecomendations(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO: вывод рекомендаций
}

func (t *TGBot) getTVsPopular(ctx context.Context, chatID int64, userData UserData) {
	m, err := t.api.GetTVPopular(ctx, userData.popularTVsPage)
	if err != nil {
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        err.Error(),
			ReplyMarkup: userData.replyKeyboard,
		})
	}

	kb := inline.New(t.bot).
		Row().
		Button("Назад", []byte("prev"), t.onTVsPopularPage).
		Button("Далее", []byte("next"), t.onTVsPopularPage)

	t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        m.Print(fmt.Sprintf("Популярные сериалы #%d", userData.popularTVsPage)),
		ReplyMarkup: kb,
	})
}

func (t *TGBot) onTVsPopular(ctx context.Context, b *bot.Bot, update *models.Update) {
	if entry, ok := t.userData[update.Message.From.ID]; ok {
		t.getTVsPopular(ctx, update.Message.Chat.ID, entry)
	}
}

func (t *TGBot) onTVsPopularPage(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	if entry, ok := t.userData[mes.Message.From.ID]; ok {
		entry.popularTVsPage = utils.HandlePage(entry.popularTVsPage, string(data))
		t.userData[mes.Message.From.ID] = entry
		t.getTVsPopular(ctx, mes.Message.Chat.ID, entry)
	}
}

func (t *TGBot) getTVsTop(ctx context.Context, chatID int64, userData UserData) {
	m, err := t.api.GetTVTop(ctx, userData.topTVsPage)
	if err != nil {
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        err.Error(),
			ReplyMarkup: userData.replyKeyboard,
		})
	}

	kb := inline.New(t.bot).
		Row().
		Button("Назад", []byte("prev"), t.onTVsTopPage).
		Button("Далее", []byte("next"), t.onTVsTopPage)

	t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        m.Print(fmt.Sprintf("Лучшие сериалы #%d", userData.topTVsPage)),
		ReplyMarkup: kb,
	})
}

func (t *TGBot) onTVsTop(ctx context.Context, b *bot.Bot, update *models.Update) {
	if entry, ok := t.userData[update.Message.From.ID]; ok {
		t.getTVsTop(ctx, update.Message.Chat.ID, entry)
	}
}

func (t *TGBot) onTVsTopPage(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	if entry, ok := t.userData[mes.Message.From.ID]; ok {
		entry.topTVsPage = utils.HandlePage(entry.topTVsPage, string(data))
		t.userData[mes.Message.From.ID] = entry
		t.getTVsTop(ctx, mes.Message.Chat.ID, entry)
	}
}

func (t *TGBot) onTVsViewed(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO: вывод просмотренных
}

func (t *TGBot) handlerReplyKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	if entry, ok := t.userData[update.Message.From.ID]; ok {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Выберите тип контента, который хотите посмотреть",
			ReplyMarkup: entry.replyKeyboard,
		})
	}
}
