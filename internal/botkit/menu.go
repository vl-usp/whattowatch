package botkit

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"whattowatch/internal/types"
	"whattowatch/internal/utils"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/inline"
	"github.com/go-telegram/ui/slider"
)

// TODO: продумать параметры внутренних функций и способы передачи данных между функциями
// для слайдеров возможно нужно будет использовать контекст
func (t *TGBot) menuHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "menuHandler", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
	log.Debug("handler func start log")

	userID := strconv.Itoa(int(update.Message.From.ID))

	menu := inline.New(b).
		Row().
		Button("Рекомендации", []byte("my:"+userID), t.onSelectRecomendations).
		Button("Общие рекомендации", []byte("joint:"+userID), t.onSelectRecomendations).
		Row().
		Button("Мои избранные", []byte(userID), t.onSelectShowFavorites)

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Выберите опцию",
		ReplyMarkup: menu,
	})
	if err != nil {
		log.Error("failed to send message", "error", err.Error())
	}
}

func (t *TGBot) onSelectRecomendations(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	log := t.log.With("fn", "onSelectMyRecomendations", "user_id", mes.Message.From.ID, "chat_id", mes.Message.Chat.ID)
	params := strings.Split(string(data), ":")
	userIDStr := params[1]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Error("failed to get user", "error", err.Error())
	}
	log.Debug("handler func start log", "user", userID)

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: mes.Message.Chat.ID,
		Text:   "Выберите тип контента:",
		ReplyMarkup: inline.New(b).
			Row().
			Button("Фильмы", []byte("MovieContentType:"+userIDStr), t.onSelectContentType).
			Button("Сериалы", []byte("TVContentType:"+userIDStr), t.onSelectContentType),
	})
	if err != nil {
		log.Error("failed to send message", "error", err.Error())
	}
}

func (t *TGBot) onSelectContentType(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	log := t.log.With("fn", "onSelectContentType", "user_id", mes.Message.From.ID, "chat_id", mes.Message.Chat.ID)
	params := strings.Split(string(data), ":")
	selectedContentType, err := types.ParseContentType(params[0])
	if err != nil {
		log.Error("failed to parse content type", "error", err.Error())
	}
	userIDStr := params[1]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Error("failed to get user", "error", err.Error())
	}
	log.Debug("handler func start log", "user", userID)

	recs, err := t.api.GetRecomendations(ctx, userID, selectedContentType)
	if err != nil {
		log.Error("failed to get recommendations", "error", err.Error())
	}

	for genre, items := range recs {
		slides := make([]slider.Slide, 0, len(items))
		for _, rec := range items {
			s := slider.Slide{
				Text:  getSliderText(rec, "Выборка по жанру: "+genre),
				Photo: t.cfg.Urls.TMDbImageUrl + rec.PosterPath,
			}
			slides = append(slides, s)
		}

		opts := []slider.Option{
			slider.OnSelect("Получить команду на добавление", false, t.sliderOnAddFavorite),
		}
		sl := slider.New(slides, opts...)
		_, err := sl.Show(ctx, b, int(mes.Message.Chat.ID))
		if err != nil {
			log.Error("failed to show slider", "error", err.Error())
		}
	}
}

func (t *TGBot) onSelectShowFavorites(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	log := t.log.With(
		"fn", "onSelectShowFavorites",
		"message", mes.Message,
		"user_id", mes.Message.From.ID,
		"chat_id", mes.Message.Chat.ID,
	)
	userID, err := strconv.Atoi(string(data))
	if err != nil {
		log.Error("failed to get user", "error", err.Error())
	}
	log.Debug("handler func start log", "user", userID)

	favoritesMap, err := t.storer.GetUserFavoritesByType(ctx, userID)
	if err != nil {
		log.Error("failed to get user favorites by type", "error", err.Error())
	}

	for _, favorites := range favoritesMap {
		slides := make([]slider.Slide, 0, len(favorites))

		for _, fav := range favorites {
			fct := types.ContentType(fav.ContentTypeID)
			var filmContentTypeName string
			switch fct {
			case types.MovieContentType:
				filmContentTypeName = "Фильмы"
			case types.TVContentType:
				filmContentTypeName = "Сериалы"
			}
			s := slider.Slide{
				Text:  getSliderText(fav, filmContentTypeName),
				Photo: t.cfg.Urls.TMDbImageUrl + fav.PosterPath,
			}
			slides = append(slides, s)
		}

		opts := []slider.Option{
			slider.OnSelect("Получить команду на удаление", false, t.sliderOnDeleteFavorite),
		}
		sl := slider.New(slides, opts...)
		_, err := sl.Show(ctx, b, int(mes.Message.Chat.ID))
		if err != nil {
			log.Error("failed to show slider", "error", err.Error())
		}
	}
}

// TODO: педелать, чтобы удалялось и добавлялось сразу, но надо узнать id пользователя

func (t *TGBot) sliderOnDeleteFavorite(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, item int) {
	log := t.log.With(
		"fn", "sliderOnDeleteFavorite",
		"message", mes.Message,
		"ctx", ctx,
		"user_id", mes.Message.From.ID,
		"chat_id", mes.Message.Chat.ID,
	)
	log.Debug("handler func start log")
	names := regexp.MustCompile(`Название:\s*(.*?)\nЖанры`).FindStringSubmatch(mes.Message.Caption)
	if names != nil {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: mes.Message.Chat.ID,
			Text:   fmt.Sprintf("Введите комманду:\n/remove %s\n", names[1]),
		})
		if err != nil {
			log.Error("failed to send message", "error", err.Error())
		}
	}
	t.removeFavoriteHandler(ctx, b, &models.Update{Message: mes.Message})
}

func (t *TGBot) sliderOnAddFavorite(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, item int) {
	log := t.log.With(
		"fn", "sliderOnDeleteFavorite",
		"ctx", ctx,
		"message", mes.Message,
		"user_id", mes.Message.From.ID,
		"chat_id", mes.Message.Chat.ID,
	)
	log.Debug("handler func start log")
	names := regexp.MustCompile(`Название:\s*(.*?)\nЖанры`).FindStringSubmatch(mes.Message.Caption)
	if names != nil {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: mes.Message.Chat.ID,
			Text:   fmt.Sprintf("Введите комманду:\n/add %s\n", names[1]),
		})
		if err != nil {
			log.Error("failed to send message", "error", err.Error())
		}
	}
	t.addFavoriteHandler(ctx, b, &models.Update{Message: mes.Message})
}

func getSliderText(c types.Content, header string) string {
	text := fmt.Sprintf(
		"\n%s\n%s\nНазвание: %s\nЖанры: %s\nОписание: %s\nДата выхода: %s\n",
		header,
		"============================",
		c.Title,
		c.Genres.String(),
		c.Overview,
		c.ReleaseDate.Time.Format("02.01.2006"),
	)

	return utils.EscapeString(text)
}
