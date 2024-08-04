package botkit

import (
	"context"
	"database/sql"
	"time"
	"whattowatch/internal/types"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (t *TGBot) registerHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	err := t.storer.InsertUser(ctx, types.User{
		ID:           update.Message.From.ID,
		FirstName:    update.Message.From.FirstName,
		LastName:     update.Message.From.LastName,
		Username:     update.Message.From.Username,
		LanguageCode: update.Message.From.LanguageCode,
		CreatedAt:    sql.NullTime{Time: time.Now(), Valid: true},
	})
	if err != nil {
		t.log.Error("error inserting user into db", "error", err)
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Произошла ошибка. Попробуйте позже.",
		})
		if err != nil {
			t.log.Error("error sending error message", "error", err)
		}
		return
	}
	// t.addFavoritesHandler(ctx, b, update)
}
