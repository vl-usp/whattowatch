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
	"github.com/go-telegram/ui/slider"
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
		Text:   "/start - регистрация\n/menu - открыть меню\n/help - помощь",
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
	titles := make([]string, 0)
	for _, title := range strings.Split(titlesStr, ",") {
		titles = append(titles, strings.TrimSpace(title))
	}

	res, err := t.api.SearchByTitles(ctx, titles)
	if err != nil {
		log.Error("failed to get movies", "error", err.Error())
		t.sendErrorMessage(ctx, update.Message.Chat.ID)
		return
	}

	opts := []slider.Option{
		slider.WithPrefix("slider_movie_search"),
	}
	sl := t.generateSlider(res, opts)
	sl.Show(ctx, b, update.Message.Chat.ID)
}

func (t *TGBot) searchByIDHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "searchHandler", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
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
		contentItem, err = t.api.GetMovie(ctx, id)
	case "/t":
		contentItem, err = t.api.GetTV(ctx, id)
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

	kb := inline.New(b).Row()
	serializedItem := types.SerializeContentItemKey(contentItem)
	if cs.IsFavorite {
		kb = kb.Button("Удалить из избранных", []byte(serializedItem), t.onRemoveFavorite)
	} else {
		kb = kb.Button("Добавить в избранные", []byte(serializedItem), t.onAddFavorite)
	}

	if cs.IsViewed {
		kb = kb.Button("Удалить из просмотренных", []byte(serializedItem), t.onRemoveViewed)
	} else {
		kb = kb.Button("Добавить в просмотренные", []byte(serializedItem), t.onAddViewed)
	}

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

func (t *TGBot) onAddFavorite(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	log := t.log.With("fn", "onAddFavorite", "user_id", mes.Message.From.ID, "chat_id", mes.Message.Chat.ID)

	item, err := types.UnserializeContentItemKey(string(data))
	if err != nil {
		log.Error("failed to unserialize item", "error", err.Error())
		t.sendErrorMessage(ctx, mes.Message.Chat.ID)
		return
	}

	err = t.storer.AddContentItemToFavorite(ctx, mes.Message.Chat.ID, item)
	if err != nil {
		log.Error("failed to add favorite", "error", err.Error())
		t.sendErrorMessage(ctx, mes.Message.Chat.ID)
	}
}

func (t *TGBot) onRemoveFavorite(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	log := t.log.With("fn", "onRemoveFavorite", "user_id", mes.Message.From.ID, "chat_id", mes.Message.Chat.ID)

	item, err := types.UnserializeContentItemKey(string(data))
	if err != nil {
		log.Error("failed to unserialize item", "error", err.Error())
		t.sendErrorMessage(ctx, mes.Message.Chat.ID)
		return
	}

	err = t.storer.RemoveContentItemFromFavorite(ctx, mes.Message.Chat.ID, item)
	if err != nil {
		log.Error("failed to remove favorite", "error", err.Error())
		t.sendErrorMessage(ctx, mes.Message.Chat.ID)
	}
}

func (t *TGBot) onAddViewed(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	log := t.log.With("fn", "onAddViewed", "user_id", mes.Message.From.ID, "chat_id", mes.Message.Chat.ID)

	item, err := types.UnserializeContentItemKey(string(data))
	if err != nil {
		log.Error("failed to unserialize item", "error", err.Error())
		t.sendErrorMessage(ctx, mes.Message.Chat.ID)
		return
	}

	err = t.storer.AddContentItemToViewed(ctx, mes.Message.Chat.ID, item)
	if err != nil {
		log.Error("failed to add viewed", "error", err.Error())
		t.sendErrorMessage(ctx, mes.Message.Chat.ID)
	}
}

func (t *TGBot) onRemoveViewed(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	log := t.log.With("fn", "onRemoveViewed", "user_id", mes.Message.From.ID, "chat_id", mes.Message.Chat.ID)

	item, err := types.UnserializeContentItemKey(string(data))
	if err != nil {
		log.Error("failed to unserialize item", "error", err.Error())
		t.sendErrorMessage(ctx, mes.Message.Chat.ID)
		return
	}

	err = t.storer.RemoveContentItemFromViewed(ctx, mes.Message.Chat.ID, item)
	if err != nil {
		log.Error("failed to remove viewed", "error", err.Error())
		t.sendErrorMessage(ctx, mes.Message.Chat.ID)
	}
}
