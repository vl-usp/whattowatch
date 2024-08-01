package tgbot

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/inline"
)

func (t *TGBot) helpHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	text := fmt.Sprintf(
		"%s\n%s\n%s\n",
		"",
		"/start - открыть меню",
		"/help - помощь",
	)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: models.ParseModeMarkdown,
	})
}

func (t *TGBot) defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	menu := inline.New(b).
		Row().
		Button("Мои рекомендации", []byte("1-1"), onInlineKeyboardSelect).
		Button("Совместные рекомендации", []byte("1-2"), onInlineKeyboardSelect).
		Row().
		Button("Добавить друга", []byte("2-1"), onInlineKeyboardSelect).
		Button("Посмотреть избранные", []byte("2-2"), onInlineKeyboardSelect).
		Button("Добавить в избранные", []byte("2-3"), onInlineKeyboardSelect).
		Row().
		Button("Отменить", []byte("cancel"), onInlineKeyboardSelect)

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Select the variant",
		ReplyMarkup: menu,
	})
	if err != nil {
		t.log.Error("error sending message", "error", err)
	}
}

func onInlineKeyboardSelect(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: mes.Message.Chat.ID,
		Text:   "Я знаю, что вы пикнули кнопку: " + string(data),
	})
}
