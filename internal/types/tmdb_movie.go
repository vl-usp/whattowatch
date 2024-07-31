package types

type TMDbMovie struct {
	ID          int
	Title       string
	Overview    string
	Popularity  float32
	PosterPath  string
	ReleaseDate string
	Budget      uint32
	Revenue     uint32
	Runtime     uint32
	VoteAverage float32
	VoteCount   uint32
}
