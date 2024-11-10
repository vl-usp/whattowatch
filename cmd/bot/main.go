package main

import (
	"log/slog"
	"os"
	"whattowatch/internal/api"
	"whattowatch/internal/botkit"
	"whattowatch/internal/config"
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

	log, file := logger.SetupLogger(cfg.Env, cfg.LogDir+"/bot")
	defer file.Close()

	log.Info("Current IP: " + utils.GetMyIP())

	postgresDB, err := postgresql.New(cfg, log)
	if err != nil {
		log.Error("storage create error", "error", err.Error())
		panic("storage create error: " + err.Error())
	}

	api, err := api.New(cfg, log)
	if err != nil {
		log.Error("API create error", "error", err.Error())
		panic("API create error: " + err.Error())
	}

	bot, err := botkit.New(cfg, log, postgresDB, api)
	if err != nil {
		log.Error("TGBot create error", "error", err.Error())
		panic("TGBot create error: " + err.Error())
	}
	bot.Start()
}
