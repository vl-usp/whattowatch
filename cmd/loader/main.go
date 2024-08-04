package main

import (
	"context"
	"whattowatch/internal/config"
	"whattowatch/internal/services/loader"
	"whattowatch/internal/storage"
	"whattowatch/pkg/logger"
)

func main() {
	cfg := config.MustLoad()

	log, file := logger.SetupLogger(cfg.Env, cfg.LogDir+"/loader")
	defer file.Close()

	storer, err := storage.New(cfg, log)
	if err != nil {
		log.Error("creating a storage error", "error", err.Error())
	}
	loader, err := loader.NewTMDbLoader(cfg.Tokens.TMDb, cfg.Urls.TMDbApiUrl, log, storer)
	if err != nil {
		log.Error("creating a loader error", "error", err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = loader.Load(ctx)
	if err != nil {
		log.Error("load error", "error", err.Error())
	}
}
