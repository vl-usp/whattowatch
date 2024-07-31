package parser

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"whattowatch/internal/storage"
	"whattowatch/internal/types"

	"github.com/PuerkitoBio/goquery"
)

type TMDBParser struct {
	Parser
}

func newTMDBParser(rawURL string, log *slog.Logger, storage storage.IStorage) (*TMDBParser, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return &TMDBParser{
		Parser: Parser{
			log:     log,
			storage: storage,
			BaseUrl: u,
			wg:      &sync.WaitGroup{},
		},
	}, nil
}

func (p *TMDBParser) Parse() error {
	ctx := context.Background()
	err := p.parseLinksWorkerPool(ctx, 1)
	if err != nil {
		return err
	}
	p.wg.Wait()

	// c := colly.NewCollector(
	// 	colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36"),
	// )

	// TODO: parse every link in worker pool
	// for _, link := range *links {
	// 	m, err := p.parseMovie(c, link.MovieUrl)
	// 	if err != nil {
	// 		p.log.Error("parse movie error", "err", err.Error())
	// 		return err
	// 	}
	// 	err = p.storage.InsertMovies(context.Background(), m)
	// }

	return err
}

// ParseLinks parse source link from tmdb and save it to storage
func (p *TMDBParser) parseLinksWorkerPool(ctx context.Context, numWorker int) error {
	pageFrom := 1
	pageTo := 50
	pageCh := make(chan int)
	resCh := make(chan types.SourceLinkMap)
	// p.wg.Add(1)
	// go func() {
	p.log.Info("start parse links", "from", pageFrom, "to", pageTo, "numWorker", numWorker)
	for i := 0; i < numWorker; i++ {
		go p.parseLinksFromPage(ctx, pageCh, resCh)
	}
	// p.wg.Done()
	// }()

	// write to pageCh
	// p.wg.Add(1)
	go func() {
		defer close(pageCh)
		for page := pageFrom; page <= pageTo; page++ {
			p.wg.Add(1)
			p.log.Debug("add page to queue", "page", pageFrom, "to", pageTo, "page", page)
			pageCh <- page
		}
		// p.wg.Done()
	}()

	// read from resCh
	// p.wg.Add(1)
	// go func() {
	for res := range resCh {
		p.log.Debug("inserting source links to storage", "links", res)
		err := p.storage.InsertSourceLinks(context.Background(), res)
		if err != nil {
			p.log.Error("failed to insert source links", "map", res, "err", err.Error())
		}
	}
	// p.wg.Done()
	// }()

	return nil
}

// func (p *TMDBParser) parseMovie(c *colly.Collector, url string) (*types.Movie, error) {
// 	u := p.GetBaseUrl()
// 	u.Path = url
// 	c.OnHTML("div.header.first", func(e *colly.HTMLElement) {
// 		// parse genres
// 		genres := make([]types.Genre, 0)
// 		e.ForEach("div.header_poster_wrapper span.genres a", func(_ int, e *colly.HTMLElement) {
// 			genres = append(genres, types.Genre{
// 				Name: e.Text,
// 			})
// 		})
// 		p.log.Debug("genres", "count", len(genres), "genres", genres)

// 		movie := &types.Movie{
// 			SourceID:    1, //TODO make it dynamic
// 			Title:       e.ChildText("div.header_poster_wrapper h2 a"),
// 			Description: e.ChildText("p"),
// 			Genres:      genres,
// 			Runtime:     e.ChildText("span.runtime"),
// 		}
// 		err := p.storage.InsertMovie(context.Background(), movie)
// 		if err != nil {
// 			p.log.Error(err.Error())
// 		}
// 	})
// 	err := c.Visit(u.String())
// 	if err != nil {
// 		return err
// 	}

// }

func (p *TMDBParser) parseLinksFromPage(_ context.Context, pageCh chan int, resCh chan types.SourceLinkMap) {
	for page := range pageCh {
		params := url.Values{
			"certification_country": []string{"RU"},
			"page":                  []string{strconv.Itoa(page)},
			"release_date.lte":      []string{time.Now().Format("2006-01-02")},
			"show_me":               []string{"everything"},
			"sort_by":               []string{"popularity.desc"},
			"vote_average.gte":      []string{"0"},
			"vote_average.lte":      []string{"10"},
			"vote_count.gte":        []string{"0"},
			"watch_region":          []string{"RU"},
			"with_runtime.gte":      []string{"0"},
			"with_runtime.lte":      []string{"400"},
			"language":              []string{"ru-RU"},
		}
		p.log.Debug("request to tmdb: ", "url", p.BaseUrl.String(), "prams", params)
		r, err := http.PostForm(p.BaseUrl.String(), params)
		if err != nil {
			p.log.Error("request error: ", "url", p.BaseUrl.String(), "prams", params)
		}
		p.log.Debug("response from tmdb: ", "url", p.BaseUrl.String(), "prams", params, "status", r.Status)

		doc, err := goquery.NewDocumentFromReader(r.Body)
		if err != nil {
			p.log.Error("error parsing response: ", "url", p.BaseUrl.String(), "prams", params)
		}
		links := make(types.SourceLinkMap, 20)
		// Find the review items
		doc.Find("h2 a").Each(func(i int, s *goquery.Selection) {
			// For each item found, get the title
			href, ok := s.Attr("href")
			if !ok {
				p.log.Error("href not found")
			}
			re := regexp.MustCompile("[0-9]+")
			id, err := strconv.Atoi(re.FindString(href))
			if err != nil {
				p.log.Error("error parsing id", "err", err.Error())
			}
			title := s.Text()
			link, _, _ := strings.Cut(href, "?language=ru-RU")
			l := types.SourceLink{
				SourceID:   p.ID,
				OriginalID: id,
				Title:      title,
				Page:       page,
				MovieUrl:   link,
			}
			p.log.Debug("found link", "link", l)
			links[id] = l
		})

		resCh <- links
		err = r.Body.Close()
		if err != nil {
			p.log.Error("error closing response: ", "url", p.BaseUrl.String(), "prams", params)
		}
		// time.Sleep(1500 * time.Millisecond)
		p.wg.Done()
	}
}

// func (p *TMDBParser) parseMovie(e *colly.HTMLElement) (*types.Movie, error) {
// 	return nil, nil
// }
