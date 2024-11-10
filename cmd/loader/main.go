package main

import (
	"context"
	"log/slog"
	"os"
	"whattowatch/internal/config"
	"whattowatch/internal/loader"
	"whattowatch/internal/storage/postgresql"
	"whattowatch/internal/utils"
	"whattowatch/pkg/logger"
)

func main() {
	cfg, err := config.MustLoad()
	if err != nil {
		ir, _ := os.Getwd()
		slog.Error("failed to load config", "error", err.Error(), "current dir", ir)
		panic("failed to load config: " + err.Error())
	}

	log, file := logger.SetupLogger(cfg.Env, cfg.LogDir+"/loader")
	defer file.Close()

	log.Info("Current IP: " + utils.GetMyIP())

	postgresDB, err := postgresql.New(cfg, log)
	if err != nil {
		log.Error("storage create error", "error", err.Error())
		panic("storage create error: " + err.Error())
	}
	loader, err := loader.NewTMDbLoader(cfg, log, postgresDB)
	if err != nil {
		log.Error("loader create error", "error", err.Error())
		panic("loader create error: " + err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = loader.Load(ctx)
	if err != nil {
		log.Error("load error", "error", err.Error())
	}
}
