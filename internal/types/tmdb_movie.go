package types

type TMDbMovie struct {
	ID          int
	Title       string
	Overview    string
	Popularity  float32
	PosterPath  string
	ReleaseDate string
	VoteAverage float32
	VoteCount   uint32
}
