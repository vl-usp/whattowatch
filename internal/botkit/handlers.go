package botkit

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (t *TGBot) helpHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "/start - регистрация\n/menu - открыть меню\n/help - помощь",
	})
	if err != nil {
		t.log.Error("error sending response for help", "error", err)
	}
}
