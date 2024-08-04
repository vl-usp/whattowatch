package config

import "os"

type Urls struct {
	TMDbApiUrl   string
	TMDbImageUrl string
}

func NewUrls() Urls {
	return Urls{
		TMDbApiUrl:   os.Getenv("TMDb_API_URL"),
		TMDbImageUrl: os.Getenv("TMDb_IMAGE_URL"),
	}
}
