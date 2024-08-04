package botkit

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (t *TGBot) getUserMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		user, _ := b.GetMe(context.Background())
		t.log.Debug("getUserMiddleware", "user", user)
		next(ctx, b, update)
	}
}
