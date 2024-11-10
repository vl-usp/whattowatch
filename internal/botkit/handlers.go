package botkit

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"
	"whattowatch/internal/types"
	"whattowatch/internal/utils"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/inline"
	"github.com/go-telegram/ui/slider"
)

type modifyUserContentFunc func(ctx context.Context, userID int64, item types.ContentItem) error
type showContentByGenreFunc func(ctx context.Context, chatID int64, userData UserData, genreID int)

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
	chatID := update.Message.Chat.ID

	log := t.log.With("fn", "searchByTitleHandler", "user_id", update.Message.From.ID, "chat_id", chatID)
	log.Debug("handler func start log")

	titlesStr := strings.Trim(update.Message.Text, "/search ")
	if titlesStr == "" {
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Введите название фильма после команды /search.\nПример: /search Начало",
		})
	}
	titles := make([]string, 0)
	for _, title := range strings.Split(titlesStr, ",") {
		titles = append(titles, strings.TrimSpace(title))
	}

	res, err := t.api.SearchByTitles(ctx, titles)
	if err != nil {
		log.Error("failed to get movies", "error", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}

	if len(res) == 0 {
		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Ничего не найдено",
		})
		return
	}

	slides := t.generateSlider(res, nil)
	_, err = slides.Show(ctx, t.bot, chatID)
	if err != nil {
		log.Error("failed to show slider", "error", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}
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

	serializedItem := types.SerializeContentItem(contentItem)
	kb := t.getContentActionKeyboard(cs, serializedItem)

	_, err = b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:      update.Message.Chat.ID,
		Photo:       &models.InputFileString{Data: contentItem.BackdropPath},
		Caption:     contentItem.GetInfo(),
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

		kb := t.getContentActionKeyboard(cs, data)

		_, err = b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID:      mes.Message.Chat.ID,
			Photo:       &models.InputFileString{Data: item.BackdropPath},
			Caption:     item.GetInfo(),
			ParseMode:   "Markdown",
			ReplyMarkup: kb,
		})
		if err != nil {
			log.Error("failed to send message", "error", err.Error())
			t.sendErrorMessage(ctx, mes.Message.Chat.ID)
		}

	}
}

func (t *TGBot) onContentByGenreHandler(fn showContentByGenreFunc, page Page) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		log := t.log.With("fn", "onContentByGenreHandler", "user_id", userID, "chat_id", chatID)
		log.Debug("handler func start log")

		genreIDStr := update.Message.Text[3:]
		genreID, err := strconv.Atoi(genreIDStr)
		if err != nil {
			log.Error("failed to parse genre id", "error", err.Error())
			t.sendErrorMessage(ctx, chatID)
			return
		}

		t.mu.RLock()
		userData, exists := t.userData[userID]
		t.mu.RUnlock()

		if !exists {
			log.Debug("user not found in userData map")
			return
		}

		userData.pagesMap[page] = 1

		t.mu.Lock()
		t.userData[userID] = userData
		t.mu.Unlock()

		fn(ctx, chatID, userData, genreID)
	}
}

func (t *TGBot) onContentGenrePageHandler(fn showContentByGenreFunc, page Page, genreID int) slider.OnCancelFunc {
	return func(ctx context.Context, b *bot.Bot, message models.MaybeInaccessibleMessage) {
		chatID := message.Message.Chat.ID

		log := t.log.With("fn", "onContentGenrePageHandler", "chat_id", chatID)
		log.Debug("handler func start log")

		t.mu.RLock()
		userData, exists := t.userData[chatID]
		t.mu.RUnlock()

		if !exists {
			log.Debug("user not found in userData map")
			return
		}

		userData.pagesMap[page] = utils.HandlePage(userData.pagesMap[page], "next")

		t.mu.Lock()
		t.userData[chatID] = userData
		t.mu.Unlock()

		fn(ctx, chatID, userData, genreID)
	}
}

func (t *TGBot) showMovieByGenre(ctx context.Context, chatID int64, userData UserData, genreID int) {
	log := t.log.With("fn", "showMoviesByGenre", "chat_id", chatID)
	log.Debug("handler func start log")

	movies, err := t.api.GetMoviesByGenre(ctx, []int{genreID}, userData.pagesMap[MovieByGenre])
	if err != nil {
		log.Error("failed to get content", "error", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}

	if len(movies) == 0 {
		t.sendErrorMessage(ctx, chatID)
		return
	}

	opts := []slider.Option{
		slider.OnCancel("Показать еще", true, t.onContentGenrePageHandler(t.showMovieByGenre, MovieByGenre, genreID)),
	}
	slides := t.generateSlider(movies, opts)
	_, err = slides.Show(ctx, t.bot, chatID)
	if err != nil {
		log.Error("failed to show slider", "error", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}
}

func (t *TGBot) showTVByGenre(ctx context.Context, chatID int64, userData UserData, genreID int) {
	log := t.log.With("fn", "showTVByGenre", "chat_id", chatID)
	log.Debug("handler func start log")

	movies, err := t.api.GetTVsByGenre(ctx, []int{genreID}, userData.pagesMap[TVByGenre])
	if err != nil {
		log.Error("failed to get content", "error", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}

	if len(movies) == 0 {
		t.sendErrorMessage(ctx, chatID)
		return
	}

	opts := []slider.Option{
		slider.OnCancel("Показать еще", true, t.onContentGenrePageHandler(t.showTVByGenre, TVByGenre, genreID)),
	}
	slides := t.generateSlider(movies, opts)
	_, err = slides.Show(ctx, t.bot, chatID)
	if err != nil {
		log.Error("failed to show slider", "error", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}
}
