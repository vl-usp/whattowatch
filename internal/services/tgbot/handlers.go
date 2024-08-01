package tgbot

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/inline"
)

func (t *TGBot) startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {

}

func (t *TGBot) defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	t.log.Debug("bot default handler", "bot", b, "update", update)

	menu := inline.New(b).
		Row().
		Button("Мои рекомендации", []byte("my"), t.onSelectRecomendations).
		Button("Совместные рекомендации", []byte("with-firend"), t.onSelectRecomendations).
		Row().
		Button("Добавить друга", []byte("2-1"), t.onAddFriend).
		Button("Посмотреть избранные", []byte("2-2"), t.onShowFavorites).
		Button("Добавить в избранные", []byte("2-3"), t.onAddToFavorites)

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Выберите опцию",
		ReplyMarkup: menu,
	})
	if err != nil {
		t.log.Error("error sending message", "error", err)
	}
}

func (t *TGBot) onSelectRecomendations(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	// должен быть сервис рекомендаций, который будет анализировать предпочтения юзера и возвращать список рекомендаций (фильмы + сериалы)
	switch string(data) {
	case "my":
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: mes.Message.Chat.ID,
			Text:   "Тут должны быть ваши рекомендации",
		})
		if err != nil {
			t.log.Error("error sending response for my recomendations", "error", err)
		}
	case "with-firend":
		// TODO
		// выбор друга из списка друзей
		// вывод совместных рекомендаций
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: mes.Message.Chat.ID,
			Text:   "Тут должны быть совместные рекомендации",
		})
		if err != nil {
			t.log.Error("error sending response for with-firend recomendations", "error", err)
		}
	}
}

func (t *TGBot) onAddFriend(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: mes.Message.Chat.ID,
		Text:   "Тут должен быть список друзей",
	})
	if err != nil {
		t.log.Error("error sending response for add friend", "error", err)
	}
}

func (t *TGBot) onShowFavorites(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
}

func (t *TGBot) onAddToFavorites(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
}
