package cache

type Cache struct {
	Genres *Genres
}

func New() *Cache {
	return &Cache{
		Genres: NewGenres(),
	}
}
