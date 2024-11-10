package botkit

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (t *TGBot) userDataMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		log := t.log.With("fn", "userDataMiddleware")

		var id int64
		if update.CallbackQuery != nil {
			id = update.CallbackQuery.From.ID
		} else {
			id = update.Message.From.ID
		}

		t.mu.RLock()
		entry, ok := t.userData[id]
		t.mu.RUnlock()

		if !ok {
			log.Debug("init user data", "userID", id)

			t.mu.Lock()
			t.userData[id] = initUserData(t.getMainKeyboard)
			t.mu.Unlock()

			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      id,
				Text:        "Выберите тип контента",
				ReplyMarkup: entry.replyKeyboard,
			})
		}

		next(ctx, b, update)
	}
}
