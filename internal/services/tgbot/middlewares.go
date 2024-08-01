package tgbot

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (t *TGBot) showMessageWithUserID(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message != nil {
			t.log.Debug("showMessageWithUserID", "user_id", update.Message.From.ID, "text", update.Message.Text)
		}
		next(ctx, b, update)
	}
}

func (t *TGBot) showMessageWithUserName(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message != nil {
			t.log.Debug("showMessageWithUserName", "first_name", update.Message.From.FirstName, "text", update.Message.Text)
		}
		next(ctx, b, update)
	}
}
