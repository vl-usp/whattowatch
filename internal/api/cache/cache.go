package cache

type Cache struct {
	Genres struct {
		TV    *Genres
		Movie *Genres
	}
}

func New() *Cache {
	return &Cache{
		Genres: struct {
			TV    *Genres
			Movie *Genres
		}{
			TV:    NewGenres(),
			Movie: NewGenres(),
		},
	}
}
