package config

import "os"

type ParseUrls struct {
	TMDb      string
	Kinopoisk string
}

func NewParseUrls() ParseUrls {
	return ParseUrls{
		TMDb:      os.Getenv("TMDB_URL"),
		Kinopoisk: os.Getenv("KINOPOISK_URL"),
	}
}
