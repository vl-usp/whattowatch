package tmdb

import (
	"context"
	"log/slog"
	"testing"
	"whattowatch/internal/types"

	"github.com/stretchr/testify/assert"
)

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
				1114738: "Список подозреваемых",
				762441:  "Тихое место: День первый",
				746036:  "Каскадёры",
				718821:  "Смерч 2",
				748783:  "Гарфилд в кино",
				823464:  "Годзилла и Конг: Новая Империя",
				1136318: "Битва за Британию",
				520763:  "Тихое место 2",
				533535:  "Дэдпул и Росомаха",
				1115623: "На расстоянии удара",
				1086747: "Смотрители",
				1001311: "Акулы в Париже",
				1011985: "Кунг-фу Панда 4",
				704673:  "Не для слабонервных",
				1111873: "Эбигейл",
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

			res, err := a.GetContent(ctx, types.Movie, ids)
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

func Test_GetMovieTop(t *testing.T) {
	a, err := New(getConfig(), slog.Default())
	assert.NoError(t, err)

	ctx := context.Background()

	res, err := a.GetMovieTop(ctx, 1)
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

func Test_GetMoviesByGenre(t *testing.T) {
	a, err := New(getConfig(), slog.Default())
	assert.NoError(t, err)

	ctx := context.Background()

	genreIDs := []int{28, 53}

	res, err := a.GetMoviesByGenre(ctx, genreIDs, 1)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Greater(t, len(res), 0)
}
