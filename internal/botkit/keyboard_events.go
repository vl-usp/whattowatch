package botkit

import (
	"context"
	"sort"
	"whattowatch/internal/types"
	"whattowatch/internal/utils"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/slider"
)

type showContentDataFunc func(ctx context.Context, chatID int64, userData UserData)
type getUserContentIDsFunc func(ctx context.Context, userID int64, contentType types.ContentType) ([]int64, error)
type getContentByIDsFunc func(ctx context.Context, contentType types.ContentType, ids []int64) (types.Content, error)

func (t *TGBot) handlerReplyKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	t.mu.RLock()
	entry, ok := t.userData[update.Message.From.ID]
	t.mu.RUnlock()

	if ok {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Выберите тип контента",
			ReplyMarkup: entry.replyKeyboard,
		})
	}
}

func (t *TGBot) onKeyboardChangeEvent(msg string, keyboardFn keyboardFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		log := t.log.With("fn", "onKeyboardChangeEvent", "user_id", userID, "chat_id", chatID)
		log.Debug("handler func start log")

		t.mu.RLock()
		userData, exists := t.userData[userID]
		t.mu.RUnlock()
		if !exists {
			log.Debug("user not found in userData map")
			return
		}

		userData.replyKeyboard = keyboardFn()

		t.mu.Lock()
		t.userData[userID] = userData
		t.mu.Unlock()

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        msg,
			ReplyMarkup: userData.replyKeyboard,
		})
	}
}

func (t *TGBot) onContentEvent(fn showContentDataFunc, page Page) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		log := t.log.With("fn", "onContentEvent", "user_id", userID, "chat_id", chatID)
		log.Debug("handler func start log")

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

		fn(ctx, chatID, userData)
	}
}

func (t *TGBot) onContentPageEvent(fn showContentDataFunc, page Page) slider.OnCancelFunc {
	return func(ctx context.Context, b *bot.Bot, message models.MaybeInaccessibleMessage) {
		chatID := message.Message.Chat.ID

		log := t.log.With("fn", "onContentPageEvent", "chat_id", chatID)
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

		fn(ctx, chatID, userData)
	}
}

func (t *TGBot) onUserContentEvent(userContentFn getUserContentIDsFunc, getContentFn getContentByIDsFunc, contentType types.ContentType, emptyMessage string) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		log := t.log.With("fn", "onUserContentEvent", "user_id", userID, "chat_id", chatID)
		log.Debug("handler func start log")

		userContentIDs, err := userContentFn(ctx, userID, contentType)
		if err != nil {
			log.Error("failed to get user content ids", "error", err.Error())
			t.sendErrorMessage(ctx, chatID)
			return
		}

		if len(userContentIDs) == 0 {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   emptyMessage,
			})
			return
		}

		content, err := getContentFn(ctx, contentType, userContentIDs)
		if err != nil {
			log.Error("failed to get content", "error", err.Error())
			t.sendErrorMessage(ctx, chatID)
			return
		}

		slides := t.generateSlider(content, nil)
		_, err = slides.Show(ctx, t.bot, chatID)
		if err != nil {
			log.Error("failed to show slider", "error", err.Error())
			t.sendErrorMessage(ctx, chatID)
			return
		}
	}
}

func (t *TGBot) onRecommendationsEvent(getContentFn getContentByIDsFunc, contentType types.ContentType) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		log := t.log.With("fn", "onRecommendationsEvent", "user_id", userID, "chat_id", chatID)
		log.Debug("handler func start log")

		favoriteIDs, err := t.storer.GetFavoriteContentIDs(ctx, userID, contentType)
		if err != nil {
			log.Error("failed to get user favorites", "error", err.Error())
			t.sendErrorMessage(ctx, chatID)
			return
		}

		viewedIDs, err := t.storer.GetViewedContentIDs(ctx, userID, contentType)
		if err != nil {
			log.Error("failed to get user viewed", "error", err.Error())
			t.sendErrorMessage(ctx, chatID)
			return
		}

		recomendations, err := getContentFn(ctx, contentType, favoriteIDs)
		if err != nil {
			log.Error("failed to get recommendations", "error", err.Error())
			t.sendErrorMessage(ctx, chatID)
			return
		}

		recomendations = recomendations.RemoveByIDs(viewedIDs).RemoveDuplicates()
		sort.Slice(recomendations, func(i, j int) bool {
			return recomendations[i].Popularity > recomendations[j].Popularity
		})

		if len(recomendations) == 0 {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "У вас нет рекомендаций",
			})
			return
		}

		slides := t.generateSlider(recomendations, nil)
		_, err = slides.Show(ctx, t.bot, chatID)
		if err != nil {
			log.Error("failed to show slider", "error", err.Error())
			t.sendErrorMessage(ctx, chatID)
			return
		}
	}
}

// showMoviePopular retrieves the popular movies content from the content service and shows it
// to the user.
func (t *TGBot) showMoviePopular(ctx context.Context, chatID int64, userData UserData) {
	log := t.log.With("fn", "showMoviePopular", "chat_id", chatID)
	log.Debug("handler func start log")

	page := userData.pagesMap[MoviePopular]
	m, err := t.api.GetMoviePopular(ctx, page)
	if err != nil {
		log.Error("failed to get popular movies", "error", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}

	opts := []slider.Option{
		slider.OnCancel("Показать еще", true, t.onContentPageEvent(t.showMoviePopular, MoviePopular)),
	}
	slides := t.generateSlider(m, opts)
	_, err = slides.Show(ctx, t.bot, chatID)
	if err != nil {
		log.Error("failed to show slider", "error", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}
}

// showMovieTop retrieves the top-rated movies content from the content service and displays it to the user.
func (t *TGBot) showMovieTop(ctx context.Context, chatID int64, userData UserData) {
	page := userData.pagesMap[MovieTop]

	log := t.log.With("fn", "showMovieTop", "chat_id", chatID, "page", page)
	log.Debug("handler func start log")

	content, err := t.api.GetMovieTop(ctx, page)
	if err != nil {
		log.Error("failed to get movie top", "error", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}

	opts := []slider.Option{
		slider.OnCancel("Показать еще", true, t.onContentPageEvent(t.showMovieTop, MovieTop)),
	}

	slides := t.generateSlider(content, opts)
	_, err = slides.Show(ctx, t.bot, chatID)
	if err != nil {
		log.Error("failed to show slider", "error", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}
}

// showTVPopular gets popular TV shows and shows them to the user
func (t *TGBot) showTVPopular(ctx context.Context, chatID int64, userData UserData) {
	page := userData.pagesMap[TVPopular]

	log := t.log.With("fn", "showTVPopular", "chat_id", chatID, "page", page)
	log.Debug("handler func start log")

	content, err := t.api.GetTVPopular(ctx, page)
	if err != nil {
		log.Error("failed to get popular tvs", "error", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}

	opts := []slider.Option{
		slider.OnCancel("Показать еще", true, t.onContentPageEvent(t.showTVPopular, TVPopular)),
	}
	slides := t.generateSlider(content, opts)
	_, err = slides.Show(ctx, t.bot, chatID)
	if err != nil {
		log.Error("failed to show slider", "error", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}
}

// showTVTop retrieves the top TV shows content from the content service and shows it
// to the user.
func (t *TGBot) showTVTop(ctx context.Context, chatID int64, userData UserData) {
	page := userData.pagesMap[TVTop]

	log := t.log.With("fn", "showTVTop", "chat_id", chatID, "page", page)
	log.Debug("handler func start log")

	content, err := t.api.GetTVTop(ctx, page)
	if err != nil {
		log.Error("failed to get top tvs", "error", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}

	opts := []slider.Option{
		slider.OnCancel("Показать еще", true, t.onContentPageEvent(t.showTVTop, TVTop)),
	}
	slides := t.generateSlider(content, opts)
	_, err = slides.Show(ctx, t.bot, chatID)
	if err != nil {
		log.Error("failed to show slider", "error", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}
}

func (t *TGBot) onGetGenresEvent(contentType types.ContentType) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		log := t.log.With("fn", "onGetGenresEvent", "user_id", userID, "chat_id", chatID)
		log.Debug("handler func start log")

		genres, err := t.api.GetGenres(ctx, contentType)
		if err != nil {
			log.Error("failed to get genres", "error", err.Error())
			t.sendErrorMessage(ctx, chatID)
			return
		}

		t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      genres.GetInfo(contentType),
			ParseMode: "Markdown",
		})
	}
}
