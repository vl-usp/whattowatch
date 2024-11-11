package tmdb

import (
	"context"
	"log/slog"
	"testing"
	"whattowatch/internal/types"

	"github.com/stretchr/testify/assert"
)

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

			res, err := a.GetContent(ctx, types.TV, ids)
			assert.NoError(t, err)
			assert.NotNil(t, res)

			assert.Equal(t, len(tt.idsNamesMap), len(res))

			for _, r := range res {
				assert.Equal(t, tt.idsNamesMap[r.ID], r.Title)
			}
		})
	}
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

func Test_GetTVTop(t *testing.T) {
	a, err := New(getConfig(), slog.Default())
	assert.NoError(t, err)

	ctx := context.Background()

	res, err := a.GetTVTop(ctx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 20, len(res))
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

func Test_GetTVsByGenre(t *testing.T) {
	a, err := New(getConfig(), slog.Default())
	assert.NoError(t, err)

	ctx := context.Background()

	genreIDs := []int{37, 35}

	res, err := a.GetTVsByGenre(ctx, genreIDs, 1)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Greater(t, len(res), 0)
}
