package botkit

import (
	"whattowatch/internal/types"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/ui/keyboard/inline"
	"github.com/go-telegram/ui/keyboard/reply"
)

type keyboardFunc func() *reply.ReplyKeyboard

func (t *TGBot) getMainKeyboard() *reply.ReplyKeyboard {
	rk := reply.New(
		t.bot,
		reply.WithPrefix("rk_main"),
		reply.IsSelective(),
		reply.ResizableKeyboard(),
	).
		Button("Фильмы 🎥", t.bot, bot.MatchTypeExact, t.onKeyboardChangeEvent("Фильмы. Выберите раздел", t.getMoviesKeyboard)).
		Row().
		Button("Сериалы 📺", t.bot, bot.MatchTypeExact, t.onKeyboardChangeEvent("Сериалы. Выберите раздел", t.getTVsKeyboard))

	return rk
}

func (t *TGBot) getMoviesKeyboard() *reply.ReplyKeyboard {
	rk := reply.New(
		t.bot,
		reply.WithPrefix("rk_movies"),
		reply.IsSelective(),
		reply.ResizableKeyboard(),
	).
		Button("Популярные 🎥", t.bot, bot.MatchTypeExact, t.onContentEvent(t.showMoviePopular, MoviePopular)).
		Button("Лучшие 🎥", t.bot, bot.MatchTypeExact, t.onContentEvent(t.showMovieTop, MovieTop)).
		Button("Жанры 🎥", t.bot, bot.MatchTypePrefix, t.onGetGenresEvent(types.Movie)).
		Row().
		Button("Рекомендации 🎥", t.bot, bot.MatchTypeExact, t.onRecommendationsEvent(t.api.GetRecommendations, types.Movie)).
		Button("Избранные 🎥", t.bot, bot.MatchTypeExact, t.onUserContentEvent(t.storer.GetFavoriteContentIDs, t.api.GetContent, types.Movie, "У вас нет избранных фильмов")).
		Button("Просмотренные 🎥", t.bot, bot.MatchTypeExact, t.onUserContentEvent(t.storer.GetViewedContentIDs, t.api.GetContent, types.Movie, "У вас нет просмотренных фильмов")).
		Row().
		Button("🔙 Назад", t.bot, bot.MatchTypePrefix, t.onKeyboardChangeEvent("Выберите тип контента", t.getMainKeyboard))

	return rk
}

func (t *TGBot) getTVsKeyboard() *reply.ReplyKeyboard {
	rk := reply.New(
		t.bot,
		reply.WithPrefix("rk_tvs"),
		reply.IsSelective(),
		reply.ResizableKeyboard(),
	).
		Button("Популярные 📺", t.bot, bot.MatchTypeExact, t.onContentEvent(t.showTVPopular, TVPopular)).
		Button("Лучшие 📺", t.bot, bot.MatchTypeExact, t.onContentEvent(t.showTVTop, TVTop)).
		Button("Жанры 📺", t.bot, bot.MatchTypePrefix, t.onGetGenresEvent(types.TV)).
		Row().
		Button("Рекомендации 📺", t.bot, bot.MatchTypeExact, t.onRecommendationsEvent(t.api.GetRecommendations, types.TV)).
		Button("Избранные 📺", t.bot, bot.MatchTypeExact, t.onUserContentEvent(t.storer.GetFavoriteContentIDs, t.api.GetContent, types.TV, "У вас нет избранных сериалов")).
		Button("Просмотренные 📺", t.bot, bot.MatchTypeExact, t.onUserContentEvent(t.storer.GetViewedContentIDs, t.api.GetContent, types.TV, "У вас нет просмотренных сериалов")).
		Row().
		Button("🔙 Назад", t.bot, bot.MatchTypePrefix, t.onKeyboardChangeEvent("Выберите тип контента", t.getMainKeyboard))

	return rk
}

func (t *TGBot) getContentActionKeyboard(contentStatus types.ContentStatus, data []byte) *inline.Keyboard {
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
