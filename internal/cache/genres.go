package cache

import "sync"

type Genres struct {
	data map[int]string
	mu   sync.RWMutex
}

func NewGenres() *Genres {
	return &Genres{
		data: make(map[int]string),
		mu:   sync.RWMutex{},
	}
}

func (c *Genres) SetGenre(key int, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func (c *Genres) GetGenre(key int) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.data[key]
	return value, ok
}

func (c *Genres) DeleteGenre(key int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

func (c *Genres) ClearGenres() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[int]string)
}
