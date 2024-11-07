package config

import "os"

type Urls struct {
	TMDbApiUrl   string
	TMDbImageUrl string
	TMDbFilesUrl string
}

func NewUrls() Urls {
	return Urls{
		TMDbApiUrl:   os.Getenv("TMDb_API_URL"),
		TMDbImageUrl: os.Getenv("TMDb_IMAGE_URL"),
		TMDbFilesUrl: os.Getenv("TMDb_FILES_URL"),
	}
}
