package loader

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
	"whattowatch/internal/config"
	"whattowatch/internal/types"

	tmdbLib "github.com/cyruzin/golang-tmdb"
	"golang.org/x/sync/errgroup"
)

type (
	Storer interface {
		InsertContent(ctx context.Context, content types.Content) error
	}

	TMDbLoader struct {
		log      *slog.Logger
		storer   Storer
		client   *tmdbLib.Client
		filesURL string
	}

	movie struct {
		Adult         bool    `json:"adult"`
		ID            int     `json:"id"`
		OriginalTitle string  `json:"original_title"`
		Popularity    float32 `json:"popularity"`
		Video         bool    `json:"video"`
	}

	tv struct {
		ID           int     `json:"id"`
		OriginalName string  `json:"original_name"`
		Popularity   float32 `json:"popularity"`
	}
)

const batchSize = 10000

func NewTMDbLoader(cfg *config.Config, logger *slog.Logger, storer Storer) (*TMDbLoader, error) {
	c, err := tmdbLib.Init(cfg.Tokens.TMDb)
	if err != nil {
		return nil, err
	}
	c.SetClientAutoRetry()

	log := logger.With("pkg", "loader")

	loader := &TMDbLoader{
		log:      log,
		storer:   storer,
		client:   c,
		filesURL: cfg.Urls.TMDbFilesUrl,
	}

	loader.log.Info("loader initialized", "url", cfg.Urls.TMDbApiUrl)

	return loader, nil
}

func (l *TMDbLoader) Load(ctx context.Context) error {
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return l.loadMovies(gCtx)
	})

	g.Go(func() error {
		return l.loadTVs(gCtx)
	})

	err := g.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (l *TMDbLoader) loadMovies(ctx context.Context) error {
	url := fmt.Sprintf("%s/movie_ids_%s.json.gz", l.filesURL, time.Now().Format("02_01_2006"))
	filepath := fmt.Sprintf("%s/%s", ".tmp", "movie_ids.json.gz")
	err := downloadFile(url, filepath)
	if err != nil {
		return err
	}

	data, err := l.readData(filepath, types.Movie)
	if err != nil {
		return err
	}

	err = l.insertData(ctx, data, batchSize)
	if err != nil {
		return err
	}

	return os.Remove(filepath)
}

func (l *TMDbLoader) loadTVs(ctx context.Context) error {
	url := fmt.Sprintf("%s/tv_series_ids_%s.json.gz", l.filesURL, time.Now().Format("02_01_2006"))
	filepath := fmt.Sprintf("%s/%s", ".tmp", "tvs_ids.json.gz")
	err := downloadFile(url, filepath)
	if err != nil {
		return err
	}

	data, err := l.readData(filepath, types.TV)
	if err != nil {
		return err
	}

	err = l.insertData(ctx, data, batchSize)
	if err != nil {
		return err
	}

	return os.Remove(filepath)
}

func downloadFile(url string, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (l *TMDbLoader) readData(filepath string, ct types.ContentType) (chan types.ContentItem, error) {
	ch := make(chan types.ContentItem, batchSize)

	go func() {
		defer close(ch)

		rawf, err := os.Open(filepath)
		if err != nil {
			l.log.Error("failed to open file", "error", err.Error())
		}
		defer rawf.Close()

		rawContents, err := gzip.NewReader(rawf)
		if err != nil {
			l.log.Error("failed to read gzip", "error", err.Error())
		}
		bufferedContents := bufio.NewReader(rawContents)

		for {
			line, err := bufferedContents.ReadBytes('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				l.log.Error("failed to read line", "error", err.Error())
			}

			switch ct {
			case types.Movie:
				var item movie
				json.Unmarshal(line[:len(line)-1], &item)
				// l.log.Info("processing line", "item", item)
				ch <- types.ContentItem{
					ID:          int64(item.ID),
					ContentType: ct,
					Title:       item.OriginalTitle,
					Popularity:  item.Popularity,
				}
			case types.TV:
				var item tv
				json.Unmarshal(line[:len(line)-1], &item)
				// l.log.Info("processing line", "item", item)
				ch <- types.ContentItem{
					ID:          int64(item.ID),
					ContentType: ct,
					Title:       item.OriginalName,
					Popularity:  item.Popularity,
				}
			}

		}
	}()

	return ch, nil
}

func (l *TMDbLoader) insertData(ctx context.Context, data chan types.ContentItem, batchSize int) error {
	var batch types.Content
	for item := range data {
		batch = append(batch, item)
		if len(batch) == batchSize {
			err := l.storer.InsertContent(ctx, batch)
			if err != nil {
				l.log.Error("failed to insert content batch", "error", err.Error())
			}
			l.log.Info("inserted batch", "items", len(batch))
			batch = types.Content{}
		}
	}
	if len(batch) > 0 {
		err := l.storer.InsertContent(ctx, batch)
		if err != nil {
			l.log.Error("failed to insert content batch", "error", err.Error())
		}
		l.log.Info("inserted batch", "items", len(batch))
	}

	return nil
}
