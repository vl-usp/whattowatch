package main

import (
	"whattowatch/internal/api"
	"whattowatch/internal/botkit"
	"whattowatch/internal/config"
	"whattowatch/internal/storage/postgresql"
	"whattowatch/pkg/logger"
)

func main() {
	cfg := config.MustLoad()

	log, file := logger.SetupLogger(cfg.Env, cfg.LogDir+"/bot")
	defer file.Close()

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
