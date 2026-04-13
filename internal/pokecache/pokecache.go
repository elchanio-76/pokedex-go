package pokecache

import (
	"sync"
	"time"
)

// Implement caching logic for pokemon data and area data

type cacheEntry struct {
	val []byte
	createdAt time.Time
}

type Cache struct {
	Entries map[string]cacheEntry
	mu sync.Mutex
}

func (c *Cache) reap(t time.Duration) {
	for {
			time.Sleep(5 * time.Millisecond)
			c.mu.Lock()
			for key, entry := range c.Entries {
				if time.Since(entry.createdAt) > t {
					delete(c.Entries, key)
				}
			}
			c.mu.Unlock()
		}
}

func NewCache(interval int) *Cache {
	c := Cache{
		Entries: make(map[string]cacheEntry),
	}
	go c.reap(time.Duration(interval) * time.Millisecond)

	return &c
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Entries[key] = cacheEntry{
		val: val,
		createdAt: time.Now().UTC(),
	}
}

func (c *Cache) Get(key string) ( []byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.Entries[key]
	if !exists {
		return nil, false
	}

	return entry.val, true
}

func (c *Cache) Delete(key string) ( []byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.Entries[key]
	if !exists {
		return nil, false
	}
	delete(c.Entries, key)
	return entry.val, true
}