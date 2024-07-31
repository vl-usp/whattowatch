package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env         string
	StorageType string
	LogDir      string
	DB          DBConfig
	ParseUrls   ParseUrls
	Tokens      Tokens
}

// MustLoad load configuration.
func MustLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := &Config{
		Env:         os.Getenv("ENV"),
		StorageType: os.Getenv("STORAGE_TYPE"),
		LogDir:      os.Getenv("LOG_DIR"),
		DB:          NewDBConfig(),
		ParseUrls:   NewParseUrls(),
		Tokens:      NewTokens(),
	}

	return cfg
}
