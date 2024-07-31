package main

import (
	"whattowatch/internal/config"
	"whattowatch/internal/services/loader"
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

	loader, err := loader.New("TMDb", cfg, log, storage)
	if err != nil {
		log.Error(err.Error())
	}

	loader.Load()
}
