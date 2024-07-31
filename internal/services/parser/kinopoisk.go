package parser

import (
	"log/slog"
	"net/url"
	"sync"
	"whattowatch/internal/storage"
)

type KinopoiskParser struct {
	Parser
}

func newKinopoiskParser(rawURL string, log *slog.Logger, storage storage.IStorage) (*KinopoiskParser, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return &KinopoiskParser{
		Parser: Parser{
			log:     log,
			storage: storage,
			BaseUrl: u,
			wg:      &sync.WaitGroup{},
		},
	}, nil
}

func (p *KinopoiskParser) Parse() error {
	return nil
}
