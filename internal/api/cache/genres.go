package cache

import "sync"

type Genres struct {
	data map[int64]string
	mu   sync.RWMutex
}

func NewGenres() *Genres {
	return &Genres{
		data: make(map[int64]string),
		mu:   sync.RWMutex{},
	}
}

func (c *Genres) Set(key int64, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func (c *Genres) Get(key int64) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.data[key]
	return value, ok
}

func (c *Genres) Delete(key int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

func (c *Genres) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[int64]string)
}
