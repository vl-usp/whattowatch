package botkit

import (
	"context"
	"database/sql"
	"strconv"
	"time"
	"whattowatch/internal/types"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/inline"
)

func (t *TGBot) registerHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "registerHandler", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
	log.Debug("handler func start log")
	err := t.storer.InsertUser(ctx, types.User{
		ID:           update.Message.From.ID,
		FirstName:    update.Message.From.FirstName,
		LastName:     update.Message.From.LastName,
		Username:     update.Message.From.Username,
		LanguageCode: update.Message.From.LanguageCode,
		CreatedAt:    sql.NullTime{Time: time.Now(), Valid: true},
	})
	if err != nil {
		log.Error("failed to insert user", "error", err.Error())
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Произошла ошибка. Попробуйте позже.",
		})
		if err != nil {
			log.Error("failed to send message", "error", err.Error())
		}
		return
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Регистрация прошла успешно",
	})
	if err != nil {
		log.Error("failed to send message", "error", err.Error())
	}
	log.Info("user registered")
}

func (t *TGBot) helpHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "helpHandler", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
	log.Debug("handler func start log")

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "/start - регистрация\n/menu - открыть меню\n/help - помощь",
	})
	if err != nil {
		log.Error("failed to send message", "error", err.Error())
	}
}

// TODO: сделать основную функцию, а эту использовать как обвязку + причесать логи
func (t *TGBot) searchMovieHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "searchMovieHandler", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
	log.Debug("handler func start log")

	strID := update.Message.Text[2:]
	id, err := strconv.Atoi(update.Message.Text[2:])
	if err != nil {
		log.Error("failed to parse id", "error", err.Error())
	}

	m, err := t.api.GetMovie(ctx, id)
	if err != nil {
		log.Error("failed to get movie", "error", err.Error())
		return
	}

	cs, err := t.storer.GetContentStatus(ctx, update.Message.From.ID, m.ID)
	if err != nil {
		log.Error("failed to get content status", "error", err.Error())
		return
	}

	kb := inline.New(b).Row()
	if cs.IsFavorite {
		kb = kb.Button("Удалить из избранных", []byte(strID), t.onRemoveFavorite)
	} else {
		kb = kb.Button("Добавить в избранные", []byte(strID), t.onAddFavorite)
	}

	if cs.IsViewed {
		kb = kb.Button("Удалить из просмотренных", []byte(strID), t.onRemoveViewed)
	} else {
		kb = kb.Button("Добавить в просмотренные", []byte(strID), t.onAddViewed)
	}

	_, err = b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:      update.Message.Chat.ID,
		Photo:       &models.InputFileString{Data: m.PosterPath},
		Caption:     m.String(),
		ReplyMarkup: kb,
	})

	if err != nil {
		log.Error("failed to send message", "error", err.Error())
	}
}

// TODO: сделать основную функцию, а эту использовать как обвязку + причесать логи
func (t *TGBot) searchTVHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "searchTVHandler", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
	log.Debug("handler func start log")

	strID := update.Message.Text[2:]
	id, err := strconv.Atoi(strID)
	if err != nil {
		log.Error("failed to parse id", "error", err.Error())
	}

	m, err := t.api.GetTV(ctx, id)
	if err != nil {
		log.Error("failed to get movie", "error", err.Error())
		return
	}

	cs, err := t.storer.GetContentStatus(ctx, update.Message.From.ID, m.ID)
	if err != nil {
		log.Error("failed to get content status", "error", err.Error())
		return
	}

	kb := inline.New(b).Row()
	if cs.IsFavorite {
		kb = kb.Button("Удалить из избранных", []byte(strID), t.onRemoveFavorite)
	} else {
		kb = kb.Button("Добавить в избранные", []byte(strID), t.onAddFavorite)
	}

	if cs.IsViewed {
		kb = kb.Button("Удалить из просмотренных", []byte(strID), t.onRemoveViewed)
	} else {
		kb = kb.Button("Добавить в просмотренные", []byte(strID), t.onAddViewed)
	}

	_, err = b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:      update.Message.Chat.ID,
		Photo:       &models.InputFileString{Data: m.PosterPath},
		Caption:     m.String(),
		ReplyMarkup: kb,
	})
	if err != nil {
		log.Error("failed to send message", "error", err.Error())
	}
}

func (t *TGBot) onAddFavorite(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	log := t.log.With("fn", "onAddFavorite", "user_id", mes.Message.From.ID, "chat_id", mes.Message.Chat.ID)

	id, err := strconv.Atoi(string(data))
	if err != nil {
		log.Error("failed to parse id", "error", err.Error())
		return
	}

	err = t.storer.AddContentToFavorite(ctx, mes.Message.Chat.ID, int64(id))
	if err != nil {
		log.Error("failed to add favorite", "error", err.Error())
	}
}

func (t *TGBot) onRemoveFavorite(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	log := t.log.With("fn", "onRemoveFavorite", "user_id", mes.Message.From.ID, "chat_id", mes.Message.Chat.ID)

	id, err := strconv.Atoi(string(data))
	if err != nil {
		log.Error("failed to parse id", "error", err.Error())
		return
	}

	err = t.storer.RemoveContentFromFavorite(ctx, mes.Message.Chat.ID, int64(id))
	if err != nil {
		log.Error("failed to remove favorite", "error", err.Error())
	}
}

func (t *TGBot) onAddViewed(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	log := t.log.With("fn", "onAddViewed", "user_id", mes.Message.From.ID, "chat_id", mes.Message.Chat.ID)

	id, err := strconv.Atoi(string(data))
	if err != nil {
		log.Error("failed to parse id", "error", err.Error())
		return
	}

	err = t.storer.AddContentToViewed(ctx, mes.Message.Chat.ID, int64(id))
	if err != nil {
		log.Error("failed to add viewed", "error", err.Error())
	}
}

func (t *TGBot) onRemoveViewed(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	log := t.log.With("fn", "onRemoveViewed", "user_id", mes.Message.From.ID, "chat_id", mes.Message.Chat.ID)

	id, err := strconv.Atoi(string(data))
	if err != nil {
		log.Error("failed to parse id", "error", err.Error())
		return
	}

	err = t.storer.RemoveContentFromViewed(ctx, mes.Message.Chat.ID, int64(id))
	if err != nil {
		log.Error("failed to remove viewed", "error", err.Error())
	}
}
