package botkit

import (
	"context"
	"sort"
	"whattowatch/internal/types"
	"whattowatch/internal/utils"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/reply"
	"github.com/go-telegram/ui/slider"
	"golang.org/x/sync/errgroup"
)

func (t *TGBot) sendErrorMessage(ctx context.Context, chatID int64) {
	t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Произошла ошибка. Попробуйте ещё раз позднее...",
	})
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
	log.Info("generating slides", "recomendations", len(content))
	slides := make([]slider.Slide, 0, len(content))

	for _, r := range content {
		slides = append(slides, slider.Slide{
			Photo: r.PosterPath,
			Text:  utils.EscapeString(r.ShortString()),
		})
	}

	if opts == nil {
		opts = []slider.Option{}
	}
	return slider.New(slides, opts...)
}

// MAIN KEYBOARD
func (t *TGBot) onMainKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	t.mu.RLock()
	entry, ok := t.userData[update.Message.From.ID]
	t.mu.RUnlock()

	if ok {
		entry.replyKeyboard = reply.New(
			b,
			reply.WithPrefix("rk_main"),
			reply.IsSelective(),
		).
			Button("Фильмы", b, bot.MatchTypeExact, t.onMoviesKeyboard).
			Row().
			Button("Сериалы", b, bot.MatchTypeExact, t.onTVsKeyboard)

		t.mu.Lock()
		t.userData[update.Message.From.ID] = entry
		t.mu.Unlock()

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Выберите тип контента, который хотите посмотреть",
			ReplyMarkup: entry.replyKeyboard,
		})
	}
}

func (t *TGBot) onMoviesKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "onMoviesKeyboard", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
	log.Debug("handler func start log")

	t.mu.RLock()
	entry, ok := t.userData[update.Message.From.ID]
	t.mu.RUnlock()

	if ok {
		entry.replyKeyboard = reply.New(
			b,
			reply.WithPrefix("rk_movies"),
			reply.IsSelective(),
		).
			Button("Рекомендации 🎥", b, bot.MatchTypeExact, t.onMoviesRecomendations).
			Button("Популярные 🎥", b, bot.MatchTypeExact, t.onMoviesPopular).
			Button("Лучшие 🎥", b, bot.MatchTypeExact, t.onMoviesTop).
			Button("Просмотренные 🎥", b, bot.MatchTypeExact, t.onMoviesViewed).
			Row().
			Button("🔙 Назад", b, bot.MatchTypePrefix, t.onMainKeyboard)

		t.mu.Lock()
		t.userData[update.Message.From.ID] = entry
		t.mu.Unlock()

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Фильмы. Выберите раздел",
			ReplyMarkup: entry.replyKeyboard,
		})
	}
}

func (t *TGBot) onTVsKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "onTVsKeyboard", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
	log.Debug("handler func start log")

	t.mu.RLock()
	entry, ok := t.userData[update.Message.From.ID]
	t.mu.RUnlock()

	if ok {
		entry.replyKeyboard = reply.New(
			b,
			reply.WithPrefix("rk_tvs"),
			reply.IsSelective(),
		).
			Button("Рекомендации 📺", b, bot.MatchTypeExact, t.onTVsRecomendations).
			Button("Популярные 📺", b, bot.MatchTypeExact, t.onTVsPopular).
			Button("Лучшие 📺", b, bot.MatchTypeExact, t.onTVsTop).
			Button("Просмотренные 📺", b, bot.MatchTypeExact, t.onTVsViewed).
			Row().
			Button("🔙 Назад", b, bot.MatchTypePrefix, t.onMainKeyboard)

		t.mu.Lock()
		t.userData[update.Message.From.ID] = entry
		t.mu.Unlock()

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Сериалы. Выберите раздел",
			ReplyMarkup: entry.replyKeyboard,
		})
	}
}

// POPULAR
// Movies popular
func (t *TGBot) onMoviesPopular(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "onMoviesPopular", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
	log.Debug("handler func start log")

	t.mu.RLock()
	entry, ok := t.userData[update.Message.From.ID]
	t.mu.RUnlock()

	if ok {
		entry.popularMoviesPage = 1

		t.mu.Lock()
		t.userData[update.Message.From.ID] = entry
		t.mu.Unlock()

		t.getMoviePopular(ctx, update.Message.Chat.ID, entry)
	}
}

func (t *TGBot) getMoviePopular(ctx context.Context, chatID int64, userData UserData) {
	log := t.log.With("fn", "getMoviePopular", "chat_id", chatID)
	m, err := t.api.GetMoviesPopular(ctx, userData.popularMoviesPage)
	if err != nil {
		log.Error("get movie popular", "err", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}

	opts := []slider.Option{
		slider.OnCancel("Показать еще", true, t.onMoviesPopularPage),
		slider.WithPrefix("slider_movie_popular"),
	}
	slides := t.generateSlider(m, opts)
	slides.Show(ctx, t.bot, chatID)
}

func (t *TGBot) onMoviesPopularPage(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage) {
	log := t.log.With("fn", "onMoviesPopularPage", "user_id", mes.Message.From.ID, "chat_id", mes.Message.Chat.ID)
	log.Debug("handler func start log")

	t.mu.RLock()
	entry, ok := t.userData[mes.Message.Chat.ID]
	t.mu.RUnlock()

	if ok {
		entry.popularMoviesPage = utils.HandlePage(entry.popularMoviesPage, "next")

		t.mu.Lock()
		t.userData[mes.Message.Chat.ID] = entry
		t.mu.Unlock()

		t.getMoviePopular(ctx, mes.Message.Chat.ID, entry)
	}
}

// TVs popular
func (t *TGBot) onTVsPopular(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "onTVsPopular", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
	log.Debug("handler func start log")

	t.mu.RLock()
	entry, ok := t.userData[update.Message.From.ID]
	t.mu.RUnlock()

	if ok {
		entry.popularTVsPage = 1

		t.mu.Lock()
		t.userData[update.Message.From.ID] = entry
		t.mu.Unlock()

		t.getTVsPopular(ctx, update.Message.Chat.ID, entry)
	}
}

func (t *TGBot) getTVsPopular(ctx context.Context, chatID int64, userData UserData) {
	log := t.log.With("fn", "getTVsPopular", "chat_id", chatID)
	m, err := t.api.GetTVPopular(ctx, userData.popularTVsPage)
	if err != nil {
		log.Error("get tv popular", "err", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}

	opts := []slider.Option{
		slider.OnCancel("Показать еще", true, t.onTVsPopularPage),
		slider.WithPrefix("slider_tv_popular"),
	}
	slides := t.generateSlider(m, opts)
	slides.Show(ctx, t.bot, chatID)
}

func (t *TGBot) onTVsPopularPage(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage) {
	t.mu.RLock()
	entry, ok := t.userData[mes.Message.From.ID]
	t.mu.RUnlock()

	if ok {
		entry.popularTVsPage = utils.HandlePage(entry.popularTVsPage, "next")

		t.mu.Lock()
		t.userData[mes.Message.From.ID] = entry
		t.mu.Unlock()

		t.getTVsPopular(ctx, mes.Message.Chat.ID, entry)
	}
}

// TOP
// Movies top
func (t *TGBot) onMoviesTop(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "onMoviesTop", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
	log.Debug("handler func start log")

	t.mu.RLock()
	entry, ok := t.userData[update.Message.From.ID]
	t.mu.RUnlock()

	if ok {
		entry.topMoviePage = 1

		t.mu.Lock()
		t.userData[update.Message.From.ID] = entry
		t.mu.Unlock()

		t.getMoviesTop(ctx, update.Message.Chat.ID, entry)
	}
}
func (t *TGBot) getMoviesTop(ctx context.Context, chatID int64, userData UserData) {
	log := t.log.With("fn", "getMoviesTop", "chat_id", chatID)
	m, err := t.api.GetMovieTop(ctx, userData.topMoviePage)
	if err != nil {
		log.Error("get movies top", "err", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}

	opts := []slider.Option{
		slider.OnCancel("Показать еще", true, t.onMoviesTopPage),
		slider.WithPrefix("slider_movie_top"),
	}
	slides := t.generateSlider(m, opts)
	slides.Show(ctx, t.bot, chatID)
}

func (t *TGBot) onMoviesTopPage(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage) {
	t.mu.RLock()
	entry, ok := t.userData[mes.Message.From.ID]
	t.mu.RUnlock()

	if ok {
		entry.topMoviePage = utils.HandlePage(entry.topMoviePage, "next")

		t.mu.Lock()
		t.userData[mes.Message.From.ID] = entry
		t.mu.Unlock()

		t.getMoviesTop(ctx, mes.Message.Chat.ID, entry)
	}
}

// TVs top
func (t *TGBot) onTVsTop(ctx context.Context, b *bot.Bot, update *models.Update) {
	t.mu.RLock()
	entry, ok := t.userData[update.Message.From.ID]
	t.mu.RUnlock()

	if ok {
		entry.topTVsPage = 1

		t.mu.Lock()
		t.userData[update.Message.From.ID] = entry
		t.mu.Unlock()

		t.getTVsTop(ctx, update.Message.Chat.ID, entry)
	}
}

func (t *TGBot) getTVsTop(ctx context.Context, chatID int64, userData UserData) {
	log := t.log.With("fn", "getTVsTop", "chat_id", chatID)
	m, err := t.api.GetTVTop(ctx, userData.topTVsPage)
	if err != nil {
		log.Error("get movies top", "err", err.Error())
		t.sendErrorMessage(ctx, chatID)
		return
	}

	opts := []slider.Option{
		slider.OnCancel("Показать еще", true, t.onTVsTopPage),
		slider.WithPrefix("slider_tv_top"),
	}
	slides := t.generateSlider(m, opts)
	slides.Show(ctx, t.bot, chatID)
}

func (t *TGBot) onTVsTopPage(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage) {
	t.mu.RLock()
	entry, ok := t.userData[mes.Message.From.ID]
	t.mu.RUnlock()

	if ok {
		entry.topTVsPage = utils.HandlePage(entry.topTVsPage, "next")

		t.mu.Lock()
		t.userData[mes.Message.From.ID] = entry
		t.mu.Unlock()

		t.getTVsTop(ctx, mes.Message.Chat.ID, entry)
	}
}

// RECOMENDATIONS
func (t *TGBot) onMoviesRecomendations(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "onMoviesRecomendations", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)

	viewedCh := make(chan []int64)
	favoriteCh := make(chan []int64)

	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		viewedIDs, err := t.storer.GetViewedContentIDs(ctx, update.Message.From.ID, types.Movie)
		if err != nil {
			log.Error("failed to get user viewed", "error", err.Error())
			return err
		}
		viewedCh <- viewedIDs
		return err
	})

	g.Go(func() error {
		favoriteIDs, err := t.storer.GetFavoriteContentIDs(ctx, update.Message.From.ID, types.Movie)
		if err != nil {
			log.Error("failed to get user favorites", "error", err.Error())
		}
		favoriteCh <- favoriteIDs
		return err
	})

	err := g.Wait()
	if err != nil {
		return
	}

	recomendations, err := t.api.GetMovieRecomendations(ctx, <-favoriteCh)
	if err != nil {
		log.Error("failed to get recomendations", "error", err.Error())
	}

	recomendations = recomendations.RemoveByIDs(<-viewedCh)
	sort.Slice(recomendations, func(i, j int) bool {
		return recomendations[i].Popularity > recomendations[j].Popularity
	})

	opts := []slider.Option{
		slider.WithPrefix("slider_movie_recomendations"),
	}
	sl := t.generateSlider(recomendations, opts)
	sl.Show(ctx, b, update.Message.Chat.ID)
}

func (t *TGBot) onTVsRecomendations(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "onTVsRecomendations", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)

	viewedCh := make(chan []int64)
	favoriteCh := make(chan []int64)

	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		viewedIDs, err := t.storer.GetViewedContentIDs(ctx, update.Message.From.ID, types.TV)
		if err != nil {
			log.Error("failed to get user viewed", "error", err.Error())
			return err
		}
		viewedCh <- viewedIDs
		return err
	})

	g.Go(func() error {
		favoriteIDs, err := t.storer.GetFavoriteContentIDs(ctx, update.Message.From.ID, types.TV)
		if err != nil {
			log.Error("failed to get user favorites", "error", err.Error())
		}
		favoriteCh <- favoriteIDs
		return err
	})

	err := g.Wait()
	if err != nil {
		return
	}

	recomendations, err := t.api.GetTVRecomendations(ctx, <-favoriteCh)
	if err != nil {
		log.Error("failed to get recomendations", "error", err.Error())
	}

	recomendations = recomendations.RemoveByIDs(<-viewedCh)
	sort.Slice(recomendations, func(i, j int) bool {
		return recomendations[i].Popularity > recomendations[j].Popularity
	})

	opts := []slider.Option{
		slider.WithPrefix("slider_tv_recomendations"),
	}
	sl := t.generateSlider(recomendations, opts)
	sl.Show(ctx, b, update.Message.Chat.ID)
}

// VIEWED
func (t *TGBot) onMoviesViewed(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "onMoviesViewed", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
	ids, err := t.storer.GetViewedContentIDs(ctx, update.Message.From.ID, types.Movie)
	if err != nil {
		log.Error("failed to get viewed", "error", err.Error())
		t.sendErrorMessage(ctx, update.Message.Chat.ID)
		return
	}

	m, err := t.api.GetMovies(ctx, ids)
	if err != nil {
		log.Error("failed to get movies", "error", err.Error())
		t.sendErrorMessage(ctx, update.Message.Chat.ID)
		return
	}

	opts := []slider.Option{
		slider.WithPrefix("slider_movie_viewed"),
	}
	sl := t.generateSlider(m, opts)
	sl.Show(ctx, b, update.Message.Chat.ID)
}

func (t *TGBot) onTVsViewed(ctx context.Context, b *bot.Bot, update *models.Update) {
	log := t.log.With("fn", "onTVsViewed", "user_id", update.Message.From.ID, "chat_id", update.Message.Chat.ID)
	ids, err := t.storer.GetViewedContentIDs(ctx, update.Message.From.ID, types.TV)
	if err != nil {
		log.Error("failed to get viewed", "error", err.Error())
		t.sendErrorMessage(ctx, update.Message.Chat.ID)
		return
	}

	m, err := t.api.GetTVs(ctx, ids)
	if err != nil {
		log.Error("failed to get movies", "error", err.Error())
		t.sendErrorMessage(ctx, update.Message.Chat.ID)
		return
	}

	opts := []slider.Option{
		slider.WithPrefix("slider_tv_viewed"),
	}
	sl := t.generateSlider(m, opts)
	sl.Show(ctx, b, update.Message.Chat.ID)
}
