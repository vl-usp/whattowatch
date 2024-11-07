package main

import (
	"whattowatch/internal/botkit"
	"whattowatch/internal/config"
	"whattowatch/internal/storage"
	"whattowatch/pkg/logger"
)

func main() {
	cfg := config.MustLoad()

	log, file := logger.SetupLogger(cfg.Env, cfg.LogDir+"/bot")
	defer file.Close()

	storer, err := storage.New(cfg, log)
	if err != nil {
		log.Error("creating a storage error", "error", err.Error())
	}

	// // Preload cache
	// cache := cache.NewGenres()
	// storer.SetGenres(cache)

	bot, err := botkit.NewTGBot(cfg, log, storer)
	if err != nil {
		log.Error("creating a TGBot error", "error", err.Error())
		panic(err.Error())
	}
	bot.Start()
}
