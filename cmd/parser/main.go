package main

import (
	"context"
	"whattowatch/internal/config"
	"whattowatch/internal/services/parser"
	"whattowatch/internal/storage"
	"whattowatch/pkg/logger"
)

func main() {
	cfg := config.MustLoad()

	log, file := logger.SetupLogger(cfg.Env, cfg.LogDir)
	defer file.Close()

	storage, err := storage.New(cfg, log)
	if err != nil {
		log.Error(err.Error())
	}

	parsers := make(parser.Parsers, 0, 2)
	sources, err := storage.GetSources(context.Background())
	if err != nil {
		log.Error("failed to get sources", "err", err.Error())
	}
	for _, source := range sources {
		p, err := parser.New(source.Name, cfg, log, storage)
		if err != nil {
			log.Error(err.Error())
		} else {
			parsers = append(parsers, p)
		}
	}

	err = parsers.ParseAll()
	if err != nil {
		log.Error("parse error", "err", err.Error())
	}
}
