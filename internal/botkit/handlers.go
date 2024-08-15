package botkit

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"whattowatch/internal/types"
	"whattowatch/internal/utils"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
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

func (t *TGBot) addFavoriteHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "addFavoriteHandler", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
	log.Debug("handler func start log")
	_, favorites, err := utils.ParseCommand(update.Message.Text)
	if err != nil {
		log.Error("failed to parse command", "error", err.Error())
	}
	filmContents, err := t.storer.GetContentByTitles(ctx, favorites)
	// TODO если некоторые фильмы не найдены, выводить сообщение, что они не найдены, так же выводить сообщение, какие фильмы были добавлены
	if err != nil {
		log.Error("failed to get film contents by titles", "error", err.Error())
	}
	err = t.storer.InsertUserFavorites(ctx, int(update.Message.From.ID), filmContents.IDs())
	if err != nil {
		log.Error("failed to insert user favorites", "error", err.Error())
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Фильмы и сериалы добавленые в избранное:\n\n%s", filmContents.PrintByContentType("Добавлен")),
	})
	if err != nil {
		log.Error("failed to send message", "error", err.Error())
	}
	t.menuHandler(ctx, b, update)
}

func (t *TGBot) removeFavoriteHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "removeFavoriteHandler", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
	log.Debug("handler func start log")
	_, favorites, err := utils.ParseCommand(update.Message.Text)
	if err != nil {
		log.Error("failed to parse command", "error", err.Error())
	}
	filmContents, err := t.storer.GetContentByTitles(ctx, favorites)
	// если некоторые фильмы не найдены, выводить сообщение, что они не найдены, так же выводить сообщение, какие фильмы были добавлены
	if err != nil {
		log.Error("failed to get film contents by titles", "error", err.Error())
	}
	err = t.storer.DeleteUserFavorites(ctx, int(update.Message.From.ID), filmContents.IDs())
	if err != nil {
		log.Error("failed to delete user favorites", "error", err.Error())
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Фильмы и сериалы удалены и избранного:\n\n%s", filmContents.PrintByContentType("Удален")),
	})
	if err != nil {
		log.Error("failed to send message", "error", err.Error())
	}
	t.menuHandler(ctx, b, update)
}
