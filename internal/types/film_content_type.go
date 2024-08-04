package types

type FilmContentType int

const (
	MovieContentType FilmContentType = iota + 1
	TVContentType
)

func (w FilmContentType) String() string {
	return [...]string{"MovieContentType", "TVContentType"}[w-1]
}

func (w FilmContentType) EnumIndex() int {
	return int(w)
}

type TMDbFilmContentType struct {
	ID   int
	Name string
}
