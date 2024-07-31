package config

import "os"

type Tokens struct {
	Kinopoisk string
	TMDb      string
	OMDb      string
}

func NewTokens() Tokens {
	return Tokens{
		Kinopoisk: os.Getenv("KINOPOISK_TOKEN"),
		TMDb:      os.Getenv("TMDb_TOKEN"),
		OMDb:      os.Getenv("OMDb_TOKEN"),
	}
}
