package parser

import (
	"fmt"
	"log/slog"
	"net/url"
	"sync"
	"whattowatch/internal/config"
	"whattowatch/internal/storage"
)

type IParser interface {
	Parse() error
	GetBaseUrl() url.URL
}

type Parser struct {
	ID      int
	BaseUrl *url.URL
	log     *slog.Logger
	storage storage.IStorage
	wg      *sync.WaitGroup
}

func (p Parser) GetBaseUrl() url.URL {
	return *p.BaseUrl
}

func New(name string, cfg *config.Config, log *slog.Logger, storage storage.IStorage) (IParser, error) {
	switch name {
	case "TMDb":
		return newTMDBParser(cfg.ParseUrls.TMDb, log, storage)
	case "Kinopoisk":
		return newKinopoiskParser(cfg.ParseUrls.Kinopoisk, log, storage)
	default:
		return nil, fmt.Errorf("unknown parser name: %s", name)
	}
}

type Parsers []IParser

func (parserList Parsers) ParseAll() error {
	for _, parser := range parserList {
		if err := parser.Parse(); err != nil {
			return err
		}
	}
	return nil
}
