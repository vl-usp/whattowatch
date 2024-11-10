package botkit

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"
	"whattowatch/internal/types"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/inline"
)

type modifyUserContentFunc func(ctx context.Context, userID int64, item types.ContentItem) error

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
		t.sendErrorMessage(ctx, update.Message.Chat.ID)
		return
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Регистрация прошла успешно",
	})
	if err != nil {
		log.Error("failed to send message", "error", err.Error())
		t.sendErrorMessage(ctx, update.Message.Chat.ID)
	}
	log.Info("user registered")
}

func (t *TGBot) helpHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "helpHandler", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
	log.Debug("handler func start log")

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "/start - Регистрация\n/menu - Открыть меню\n/search - Поиск по названию. Пример: /search Начало\n/help - Помощь",
	})
	if err != nil {
		log.Error("failed to send message", "error", err.Error())
		t.sendErrorMessage(ctx, update.Message.Chat.ID)
	}
}

func (t *TGBot) searchByTitleHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "searchByTitleHandler", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
	log.Debug("handler func start log")

	titlesStr := strings.Trim(update.Message.Text, "/search ")
	if titlesStr == "" {
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Введите название фильма после команды /search. Пример: /search Начало",
		})
	}
	titles := make([]string, 0)
	for _, title := range strings.Split(titlesStr, ",") {
		titles = append(titles, strings.TrimSpace(title))
	}

	res, err := t.content.SearchByTitles(ctx, titles)
	if err != nil {
		log.Error("failed to get movies", "error", err.Error())
		t.sendErrorMessage(ctx, update.Message.Chat.ID)
		return
	}

	if len(res) == 0 {
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Ничего не найдено",
		})
		return
	}

	sl := t.generateSlider(res, nil)
	sl.Show(ctx, b, update.Message.Chat.ID)
}

func (t *TGBot) searchByIDHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "searchByIDHandler", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
	log.Debug("handler func start log")

	contentType := update.Message.Text[:2]
	strID := update.Message.Text[2:]
	id, err := strconv.Atoi(strID)
	if err != nil {
		log.Error("failed to parse id", "error", err.Error())
		t.sendErrorMessage(ctx, update.Message.Chat.ID)
		return
	}

	var contentItem types.ContentItem
	switch contentType {
	case "/f":
		contentItem, err = t.content.GetMovie(ctx, id)
	case "/t":
		contentItem, err = t.content.GetTV(ctx, id)
	}
	if err != nil {
		log.Error("failed to get content item", "error", err.Error())
		t.sendErrorMessage(ctx, update.Message.Chat.ID)
		return
	}

	cs, err := t.storer.GetContentStatus(ctx, update.Message.From.ID, contentItem)
	if err != nil {
		log.Error("failed to get content status", "error", err.Error())
		t.sendErrorMessage(ctx, update.Message.Chat.ID)
		return
	}

	serializedItem := types.SerializeContentItem(contentItem)
	kb := t.getInlineKeyboard(cs, serializedItem)

	_, err = b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:      update.Message.Chat.ID,
		Photo:       &models.InputFileString{Data: contentItem.PosterPath},
		Caption:     contentItem.String(),
		ParseMode:   "Markdown",
		ReplyMarkup: kb,
	})
	if err != nil {
		log.Error("failed to send message", "error", err.Error())
		t.sendErrorMessage(ctx, update.Message.Chat.ID)
	}
}

func (t *TGBot) onContentActionEvent(fn modifyUserContentFunc) inline.OnSelect {
	return func(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
		chatID := mes.Message.Chat.ID

		log := t.log.With("fn", "onContentActionEvent", "chat_id", chatID)

		item, err := types.UnserializeContentItem(data)
		if err != nil {
			log.Error("failed to unserialize item", "error", err.Error())
			t.sendErrorMessage(ctx, mes.Message.Chat.ID)
			return
		}

		err = fn(ctx, chatID, item)
		if err != nil {
			log.Error("failed to modify content", "error", err.Error())
			t.sendErrorMessage(ctx, mes.Message.Chat.ID)
			return
		}

		cs, err := t.storer.GetContentStatus(ctx, chatID, item)
		if err != nil {
			log.Error("failed to get content status", "error", err.Error())
			t.sendErrorMessage(ctx, mes.Message.Chat.ID)
			return
		}

		kb := t.getInlineKeyboard(cs, data)

		_, err = b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID:      mes.Message.Chat.ID,
			Photo:       &models.InputFileString{Data: item.PosterPath},
			Caption:     item.String(),
			ParseMode:   "Markdown",
			ReplyMarkup: kb,
		})
		if err != nil {
			log.Error("failed to send message", "error", err.Error())
			t.sendErrorMessage(ctx, mes.Message.Chat.ID)
		}

	}
}

func (t *TGBot) getInlineKeyboard(contentStatus types.ContentStatus, data []byte) *inline.Keyboard {
	kb := inline.New(t.bot).Row()
	if contentStatus.IsFavorite {
		kb = kb.Button("Удалить из избранных", data, t.onContentActionEvent(t.storer.RemoveContentItemFromFavorite))
	} else {
		kb = kb.Button("Добавить в избранные", data, t.onContentActionEvent(t.storer.AddContentItemToFavorite))
	}

	if contentStatus.IsViewed {
		kb = kb.Button("Удалить из просмотренных", data, t.onContentActionEvent(t.storer.RemoveContentItemFromViewed))
	} else {
		kb = kb.Button("Добавить в просмотренные", data, t.onContentActionEvent(t.storer.AddContentItemToViewed))
	}

	return kb
}
