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

func (t *TGBot) searchMovieHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "searchMovieHandler", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
	log.Debug("handler func start log")

	id, err := strconv.Atoi(update.Message.Text[2:])
	if err != nil {
		log.Error("failed to parse id", "error", err.Error())
	}

	m, err := t.api.GetMovie(ctx, id)
	if err != nil {
		log.Error("failed to get movie", "error", err.Error())
		return
	}

	// TODO проверять есть ли в избранном и просмотренном и в завимисмоти от этого высерать разные экшены
	kb := inline.New(b).
		Row().
		Button("Добавить в избранные", []byte("add_favorite"), t.onMoviesPopularPage).
		Button("Добавить в просмотренные", []byte("add_viewed"), t.onMoviesPopularPage)

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

func (t *TGBot) searchTVHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "searchTVHandler", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
	log.Debug("handler func start log")

	id, err := strconv.Atoi(update.Message.Text[2:])
	if err != nil {
		log.Error("failed to parse id", "error", err.Error())
	}

	m, err := t.api.GetTV(ctx, id)
	if err != nil {
		log.Error("failed to get movie", "error", err.Error())
		return
	}

	// TODO проверять есть ли в избранном и просмотренном и в завимисмоти от этого высерать разные экшены
	kb := inline.New(b).
		Row().
		Button("Добавить в избранные", []byte("add_favorite"), t.onMoviesPopularPage).
		Button("Добавить в просмотренные", []byte("add_viewed"), t.onMoviesPopularPage)

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
