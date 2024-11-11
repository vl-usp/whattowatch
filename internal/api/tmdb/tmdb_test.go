package tmdb

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"whattowatch/internal/config"

	"github.com/stretchr/testify/assert"
)

func getConfig() *config.Config {
	cfg, err := config.MustLoad("../../../.env")
	if err != nil {
		fmt.Println("failed to load config: " + err.Error())
		return nil
	}
	return cfg
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
