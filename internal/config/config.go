package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BotName     string
	Env         string
	StorageType string
	LogDir      string
	DB          DBConfig
	Tokens      Tokens
	Urls        Urls
}

// MustLoad load configuration.
func MustLoad(filenames ...string) *Config {
	err := godotenv.Load(filenames...)
	if err != nil {
		slog.Error("Error loading .env file")
	}

	cfg := &Config{
		BotName:     os.Getenv("BOT_NAME"),
		Env:         os.Getenv("ENV"),
		StorageType: os.Getenv("STORAGE_TYPE"),
		LogDir:      os.Getenv("LOG_DIR"),
		DB:          NewDBConfig(),
		Tokens:      NewTokens(),
		Urls:        NewUrls(),
	}

	return cfg
}
