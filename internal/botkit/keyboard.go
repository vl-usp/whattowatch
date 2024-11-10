package botkit

import (
	"context"
	"whattowatch/internal/types"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/reply"
	"github.com/go-telegram/ui/slider"
)

type keyboardFunc func() *reply.ReplyKeyboard

func (t *TGBot) getMainKeyboard() *reply.ReplyKeyboard {
	rk := reply.New(
		t.bot,
		reply.WithPrefix("rk_main"),
		reply.IsSelective(),
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
	).
		Button("Рекомендации 🎥", t.bot, bot.MatchTypeExact, t.onRecommendationsEvent(t.content.GetMovieRecommendations, types.Movie)).
		Button("Популярные 🎥", t.bot, bot.MatchTypeExact, t.onContentEvent(t.getMoviePopular, MoviePopular)).
		Button("Лучшие 🎥", t.bot, bot.MatchTypeExact, t.onContentEvent(t.getMovieTop, MovieTop)).
		Row().
		Button("Избранные 🎥", t.bot, bot.MatchTypeExact, t.onUserContentEvent(t.storer.GetFavoriteContentIDs, t.content.GetMovies, types.Movie, "У вас нет избранных фильмов")).
		Button("Просмотренные 🎥", t.bot, bot.MatchTypeExact, t.onUserContentEvent(t.storer.GetViewedContentIDs, t.content.GetMovies, types.Movie, "У вас нет просмотренных фильмов")).
		Row().
		Button("🔙 Назад", t.bot, bot.MatchTypePrefix, t.onKeyboardChangeEvent("Выберите тип контента", t.getMainKeyboard))

	return rk
}

func (t *TGBot) getTVsKeyboard() *reply.ReplyKeyboard {
	rk := reply.New(
		t.bot,
		reply.WithPrefix("rk_tvs"),
		reply.IsSelective(),
	).
		Button("Рекомендации 📺", t.bot, bot.MatchTypeExact, t.onRecommendationsEvent(t.content.GetTVRecommendations, types.Movie)).
		Button("Популярные 📺", t.bot, bot.MatchTypeExact, t.onContentEvent(t.getTVPopular, TVPopular)).
		Button("Лучшие 📺", t.bot, bot.MatchTypeExact, t.onContentEvent(t.getTVTop, TVTop)).
		Row().
		Button("Избранные 📺", t.bot, bot.MatchTypeExact, t.onUserContentEvent(t.storer.GetFavoriteContentIDs, t.content.GetTVs, types.TV, "У вас нет избранных сериалов")).
		Button("Просмотренные 📺", t.bot, bot.MatchTypeExact, t.onUserContentEvent(t.storer.GetViewedContentIDs, t.content.GetTVs, types.TV, "У вас нет просмотренных сериалов")).
		Row().
		Button("🔙 Назад", t.bot, bot.MatchTypePrefix, t.onKeyboardChangeEvent("Выберите тип контента", t.getMainKeyboard))

	return rk
}

func (t *TGBot) handlerReplyKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	t.mu.RLock()
	entry, ok := t.userData[update.Message.From.ID]
	t.mu.RUnlock()

	if ok {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Выберите тип контента, который хотите посмотреть",
			ReplyMarkup: entry.replyKeyboard,
		})
	}
}

func (t *TGBot) generateSlider(content types.Content, opts []slider.Option) *slider.Slider {
	log := t.log.With("fn", "generateSlider")
	log.Info("generating slides", "count", len(content))

	limit := 50
	if len(content) > limit {
		log.Warn("too many slides.", "limit", limit, "count", len(content))
		content = content[:limit]
	}

	slides := make([]slider.Slide, 0, limit)

	for _, r := range content {
		// log.Debug("generating slide", "title", r.Title, "short string", r.ShortString())
		slides = append(slides, slider.Slide{
			Photo: r.PosterPath,
			Text:  r.ShortString(),
		})
	}

	// log.Debug("slides generated", "count", len(slides))

	if opts == nil {
		opts = []slider.Option{}
	}
	return slider.New(slides, opts...)
}
