package config

import "os"

type Tokens struct {
	TGBot string
	TMDb  string
}

func NewTokens() Tokens {
	return Tokens{
		TGBot: os.Getenv("TG_BOT_TOKEN"),
		TMDb:  os.Getenv("TMDb_TOKEN"),
	}
}
