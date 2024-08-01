package config

import "os"

type Tokens struct {
	TGBot     string
	Kinopoisk string
	TMDb      string
	OMDb      string
}

func NewTokens() Tokens {
	return Tokens{
		TGBot:     os.Getenv("TG_BOT_TOKEN"),
		Kinopoisk: os.Getenv("KINOPOISK_TOKEN"),
		TMDb:      os.Getenv("TMDb_TOKEN"),
		OMDb:      os.Getenv("OMDb_TOKEN"),
	}
}
