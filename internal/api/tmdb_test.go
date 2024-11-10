package api

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"whattowatch/internal/config"

	"github.com/stretchr/testify/assert"
)

func getConfig() *config.Config {
	cfg, err := config.MustLoad("../../.env")
	if err != nil {
		fmt.Println("failed to load config: " + err.Error())
		return nil
	}
	return cfg
}

func Test_GetMovie(t *testing.T) {
	tests := []struct {
		name       string
		movieID    int64
		movieTitle string
	}{
		{
			name:       "OK 1",
			movieID:    150540,
			movieTitle: "Головоломка",
		},
		{
			name:       "OK 2",
			movieID:    1022789,
			movieTitle: "Головоломка 2",
		},
		{
			name:       "OK 3",
			movieID:    787699,
			movieTitle: "Вонка",
		},
		{
			name:       "OK 4",
			movieID:    9737,
			movieTitle: "Плохие парни",
		},
		{
			name:       "OK 5",
			movieID:    974635,
			movieTitle: "Я не киллер",
		},
	}

	a, err := New(getConfig(), slog.Default())
	assert.NoError(t, err)

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := a.GetMovie(ctx, int(tt.movieID))
			assert.NoError(t, err)
			assert.NotNil(t, res)

			assert.Equal(t, tt.movieID, res.ID)
			assert.Equal(t, tt.movieTitle, res.Title)
		})
	}
}

func Test_GetMovies(t *testing.T) {
	tests := []struct {
		name        string
		idsNamesMap map[int64]string
	}{
		{
			name: "OK",
			idsNamesMap: map[int64]string{
				150540:  "Головоломка",
				1022789: "Головоломка 2",
				787699:  "Вонка",
				9737:    "Плохие парни",
				974635:  "Я не киллер",
			},
		},
	}

	a, err := New(getConfig(), slog.Default())
	assert.NoError(t, err)

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids := make([]int64, 0, len(tt.idsNamesMap))
			for k := range tt.idsNamesMap {
				ids = append(ids, k)
			}

			res, err := a.GetMovies(ctx, ids)
			assert.NoError(t, err)
			assert.NotNil(t, res)

			assert.Equal(t, len(tt.idsNamesMap), len(res))

			for _, r := range res {
				assert.Equal(t, tt.idsNamesMap[r.ID], r.Title)
			}
		})
	}
}

func Test_GetTV(t *testing.T) {
	tests := []struct {
		name    string
		tvID    int64
		tvTitle string
	}{
		{
			name:    "OK 1",
			tvID:    79744,
			tvTitle: "Новобранец",
		},
		{
			name:    "OK 2",
			tvID:    1100,
			tvTitle: "Как я встретил вашу маму",
		},
		{
			name:    "OK 3",
			tvID:    1403,
			tvTitle: "Агенты «Щ.И.Т.»",
		},
		{
			name:    "OK 4",
			tvID:    91185,
			tvTitle: "Речной катер",
		},
		{
			name:    "OK 5",
			tvID:    67136,
			tvTitle: "Это мы",
		},
	}

	a, err := New(getConfig(), slog.Default())
	assert.NoError(t, err)

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := a.GetTV(ctx, int(tt.tvID))
			assert.NoError(t, err)
			assert.NotNil(t, res)

			assert.Equal(t, tt.tvID, res.ID)
			assert.Equal(t, tt.tvTitle, res.Title)
		})
	}
}

func Test_GetTVs(t *testing.T) {
	tests := []struct {
		name        string
		idsNamesMap map[int64]string
	}{
		{
			name: "OK",
			idsNamesMap: map[int64]string{
				79744: "Новобранец",
				1100:  "Как я встретил вашу маму",
				1403:  "Агенты «Щ.И.Т.»",
				91185: "Речной катер",
				67136: "Это мы",
			},
		},
	}

	a, err := New(getConfig(), slog.Default())
	assert.NoError(t, err)

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids := make([]int64, 0, len(tt.idsNamesMap))
			for k := range tt.idsNamesMap {
				ids = append(ids, k)
			}

			res, err := a.GetTVs(ctx, ids)
			assert.NoError(t, err)
			assert.NotNil(t, res)

			assert.Equal(t, len(tt.idsNamesMap), len(res))

			for _, r := range res {
				assert.Equal(t, tt.idsNamesMap[r.ID], r.Title)
			}
		})
	}
}

func Test_GetMoviePopular(t *testing.T) {
	a, err := New(getConfig(), slog.Default())
	assert.NoError(t, err)

	ctx := context.Background()

	res, err := a.GetMoviePopular(ctx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 20, len(res))
}

func Test_GetTVPopular(t *testing.T) {
	a, err := New(getConfig(), slog.Default())
	assert.NoError(t, err)

	ctx := context.Background()

	res, err := a.GetTVPopular(ctx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 20, len(res))
}

func Test_GetMovieTop(t *testing.T) {
	a, err := New(getConfig(), slog.Default())
	assert.NoError(t, err)

	ctx := context.Background()

	res, err := a.GetMovieTop(ctx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 20, len(res))
}

func Test_GetTVTop(t *testing.T) {
	a, err := New(getConfig(), slog.Default())
	assert.NoError(t, err)

	ctx := context.Background()

	res, err := a.GetTVTop(ctx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 20, len(res))
}

func Test_GetMovieRecommendations(t *testing.T) {
	a, err := New(getConfig(), slog.Default())
	assert.NoError(t, err)

	ctx := context.Background()

	ids := []int64{150540, 1022789, 787699, 9737, 974635}

	res, err := a.GetMovieRecommendations(ctx, ids)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Greater(t, len(res), 0)
}

func Test_GetTVRecommendations(t *testing.T) {
	a, err := New(getConfig(), slog.Default())
	assert.NoError(t, err)

	ctx := context.Background()

	ids := []int64{79744, 1100, 1403, 91185, 67136}

	res, err := a.GetTVRecommendations(ctx, ids)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Greater(t, len(res), 0)
}

func Test_SearchByTitls(t *testing.T) {
	a, err := New(getConfig(), slog.Default())
	assert.NoError(t, err)

	ctx := context.Background()

	titlesIDMap := map[int64]string{
		79744: "Новобранец",
		1100:  "Как я встретил вашу маму",
		1403:  "Агенты «Щ.И.Т.»",
		91185: "Речной катер",
		67136: "Это мы",
	}

	titles := make([]string, 0, len(titlesIDMap))
	for k := range titlesIDMap {
		titles = append(titles, titlesIDMap[k])
	}

	res, err := a.SearchByTitles(ctx, titles)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Greater(t, len(res), 0)
}

func Test_cache(t *testing.T) {
	genresMap := map[int64]string{
		28:    "боевик",
		18:    "драма",
		53:    "триллер",
		10749: "мелодрама",
		14:    "фэнтези",
	}

	a, err := New(getConfig(), slog.Default())
	assert.NoError(t, err)
	assert.NotNil(t, a.cache)

	for k, v := range genresMap {
		genre, _ := a.cache.Genres.Movie.Get(k)
		assert.Equal(t, v, genre)
	}
}

func Test_GetMoviesByGenre(t *testing.T) {
	a, err := New(getConfig(), slog.Default())
	assert.NoError(t, err)

	ctx := context.Background()

	genreIDs := []int{28, 53}

	res, err := a.GetMoviesByGenre(ctx, genreIDs, 1)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Greater(t, len(res), 0)

	for _, r := range res {
		assert.Contains(t, genreIDs, r.Genres[0].ID)
	}
}

func Test_GetTVsByGenre(t *testing.T) {
	a, err := New(getConfig(), slog.Default())
	assert.NoError(t, err)

	ctx := context.Background()

	genreIDs := []int{10759, 10765}

	res, err := a.GetTVsByGenre(ctx, genreIDs, 1)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Greater(t, len(res), 0)

	for _, r := range res {
		assert.Contains(t, genreIDs, r.Genres[0].ID)
	}
}
