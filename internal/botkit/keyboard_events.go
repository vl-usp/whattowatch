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

type getRatingDataFunc func(ctx context.Context, chatID int64, userData UserData)
type getUserContentIDsFunc func(ctx context.Context, userID int64, contentType types.ContentType) ([]int64, error)
type getContentByIDsFunc func(ctx context.Context, ids []int64) (types.Content, error)

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

func (t *TGBot) onContentEvent(fn getRatingDataFunc, page Page) bot.HandlerFunc {
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

func (t *TGBot) onContentPageEvent(fn getRatingDataFunc, page Page) slider.OnCancelFunc {
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

func (t *TGBot) onUserContentEvent(emptyMessage string, userContentFn getUserContentIDsFunc, getContentFn getContentByIDsFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		log := t.log.With("fn", "onUserContentEvent", "user_id", userID, "chat_id", chatID)
		log.Debug("handler func start log")

		userContentIDs, err := userContentFn(ctx, userID, types.Movie)
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

		content, err := getContentFn(ctx, userContentIDs)
		if err != nil {
			log.Error("failed to get content", "error", err.Error())
			t.sendErrorMessage(ctx, chatID)
			return
		}

		opts := []slider.Option{}
		sl := t.generateSlider(content, opts)
		sl.Show(ctx, b, chatID)
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

		recomendations, err := getContentFn(ctx, favoriteIDs)
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

		opts := []slider.Option{}
		sl := t.generateSlider(recomendations, opts)
		sl.Show(ctx, b, chatID)
	}
}

// getMoviePopular retrieves the popular movies content from the content service and shows it
// to the user.
func (t *TGBot) getMoviePopular(ctx context.Context, chatID int64, userData UserData) {
	log := t.log.With("fn", "getMoviePopular", "chat_id", chatID)
	log.Debug("starting getMoviePopular function")

	page := userData.pagesMap[MoviePopular]
	m, err := t.content.GetMoviePopular(ctx, page)
	if err != nil {
		log.Error("failed to get popular movies", "error", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}

	opts := []slider.Option{
		slider.WithPrefix("slider_movie_popular"),
		slider.OnCancel("Показать еще", true, t.onContentPageEvent(t.getMoviePopular, MoviePopular)),
	}
	slides := t.generateSlider(m, opts)
	slides.Show(ctx, t.bot, chatID)
}

// getMoviesTop retrieves the top-rated movies content from the content service and displays it to the user.
func (t *TGBot) getMovieTop(ctx context.Context, chatID int64, userData UserData) {
	page := userData.pagesMap[MovieTop]

	log := t.log.With("fn", "getMoviesTop", "chat_id", chatID, "page", page)
	log.Debug("handler func start log")

	content, err := t.content.GetMovieTop(ctx, page)
	if err != nil {
		log.Error("failed to get movie top", "error", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}

	opts := []slider.Option{
		slider.WithPrefix("slider_movie_top"),
		slider.OnCancel("Показать еще", true, t.onContentPageEvent(t.getMovieTop, MovieTop)),
	}

	slides := t.generateSlider(content, opts)
	slides.Show(ctx, t.bot, chatID)
}

// getTVsPopular gets popular TV shows and shows them to the user
func (t *TGBot) getTVPopular(ctx context.Context, chatID int64, userData UserData) {
	page := userData.pagesMap[TVPopular]

	log := t.log.With("fn", "getTVsPopular", "chat_id", chatID, "page", page)
	log.Debug("handler func start log")

	content, err := t.content.GetTVPopular(ctx, page)
	if err != nil {
		log.Error("failed to get popular tvs", "error", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}

	opts := []slider.Option{
		slider.WithPrefix("slider_tv_popular"),
		slider.OnCancel("Показать еще", true, t.onContentPageEvent(t.getTVPopular, TVPopular)),
	}
	slides := t.generateSlider(content, opts)
	slides.Show(ctx, t.bot, chatID)
}

// getTVsTop retrieves the top TV shows content from the content service and shows it
// to the user.
func (t *TGBot) getTVTop(ctx context.Context, chatID int64, userData UserData) {
	page := userData.pagesMap[TVTop]

	log := t.log.With("fn", "getTVsTop", "chat_id", chatID, "page", page)
	log.Debug("starting getTVsTop function")

	content, err := t.content.GetTVTop(ctx, page)
	if err != nil {
		log.Error("failed to get top tvs", "error", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}

	opts := []slider.Option{
		slider.OnCancel("Показать еще", true, t.onContentPageEvent(t.getTVTop, TVTop)),
		slider.WithPrefix("slider_tv_top"),
	}
	slides := t.generateSlider(content, opts)
	slides.Show(ctx, t.bot, chatID)
}
