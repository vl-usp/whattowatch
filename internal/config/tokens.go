package config

import "os"

type Tokens struct {
	TGBotToken string
	Kinopoisk  string
	TMDb       string
	OMDb       string
}

func NewTokens() Tokens {
	return Tokens{
		TGBotToken: os.Getenv("TG_BOT_TOKEN"),
		Kinopoisk:  os.Getenv("KINOPOISK_TOKEN"),
		TMDb:       os.Getenv("TMDb_TOKEN"),
		OMDb:       os.Getenv("OMDb_TOKEN"),
	}
}
