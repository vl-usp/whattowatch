package botkit

import (
	"context"
	"regexp"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/inline"
)

func (t *TGBot) menuHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	menu := inline.New(b).
		Row().
		Button("Рекомендации", []byte("my"), t.onSelectRecomendations).
		Button("Совместные рекомендации", []byte("joint"), t.onSelectRecomendations).
		Row().
		Button("Мои избранные", []byte("favorites"), t.onShowFavorites)

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Выберите опцию",
		ReplyMarkup: menu,
	})
	if err != nil {
		t.log.Error("error sending message", "error", err)
	}
}

func (t *TGBot) addFavoriteHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	favoritesStr := regexp.MustCompile(`^\/(\w|\-)+`).ReplaceAllString(update.Message.Text, "")
	favorites := strings.Split(strings.Trim(favoritesStr, " "), ",")
	// t.log.Debug("addFavoriteMoviesHandler input", "message", update.Message.Text, "str", favoritesStr, "favorites", favorites)
	for i := 0; i < len(favorites); i++ {
		favorites[i] = strings.Trim(favorites[i], " ")
	}
	filmContents, err := t.storer.GetFilmContentByTitles(ctx, favorites)
	// если некоторые фильмы не найдены, выводить сообщение, что они не найдены, так же выводить сообщение, какие фильмы были добавлены
	t.log.Debug("addFavoriteHandler", "favorites", favorites, "film_content", filmContents)
	if err != nil {
		t.log.Error("error getting film content by names", "error", err)
	}

	err = t.storer.InsertUserFavorites(ctx, int(update.Message.From.ID), filmContents.IDs())
	if err != nil {
		t.log.Error("error inserting user favorites", "error", err)
	}
}

func (t *TGBot) onSelectRecomendations(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	// Получить из хранилища мои избранные и вернуть их. Если ничего нет, то вернуть ошибку
	// должен быть сервис рекомендаций, который будет анализировать предпочтения юзера и возвращать список рекомендаций (фильмы + сериалы)
	switch string(data) {
	case "my":
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: mes.Message.Chat.ID,
			Text:   "Тут должны быть личные рекомендации",
		})
		if err != nil {
			t.log.Error("error sending response for with-firend recomendations", "error", err)
		}
	case "joint":
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: mes.Message.Chat.ID,
			Text:   "Тут должны быть совместные рекомендации",
		})
		if err != nil {
			t.log.Error("error sending response for with-firend recomendations", "error", err)
		}
	}
}

func (t *TGBot) onShowFavorites(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	// movies, tvs, err := t.storer.GetUserFavorites(ctx, int(mes.Message.From.ID))
	// t.log.Debug("onShowFavorites", "movies", movies, "tvs", tvs)
	// if err != nil {
	// 	t.log.Error("error getting favorites", "error", err)
	// }

}
