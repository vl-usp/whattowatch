package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BotName string
	Env     string
	LogDir  string
	DB      DBConfig
	Tokens  Tokens
	Urls    Urls
}

// MustLoad load configuration.
func MustLoad(filenames ...string) (*Config, error) {
	err := godotenv.Load(filenames...)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		BotName: os.Getenv("BOT_NAME"),
		Env:     os.Getenv("ENV"),
		LogDir:  os.Getenv("LOG_DIR"),
		DB:      NewDBConfig(),
		Tokens:  NewTokens(),
		Urls:    NewUrls(),
	}

	return cfg, nil
}
