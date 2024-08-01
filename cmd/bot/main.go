package main

import (
	"whattowatch/internal/config"
	"whattowatch/internal/services/tgbot"
	"whattowatch/pkg/logger"
)

// Send any text message to the bot after the bot has been started
func main() {
	cfg := config.MustLoad()

	log, file := logger.SetupLogger(cfg.Env, cfg.LogDir+"/bot")
	defer file.Close()

	bot, err := tgbot.New(cfg, log)
	if err != nil {
		log.Error("creating a TGBot error", "error", err.Error())
		panic(err.Error())
	}
	bot.Start()
}
