package geo

import (
	"sync"
	"time"
)

type cacheEntry struct {
	data      []byte
	expiresAt time.Time
}

type ttlCache struct {
	mu    sync.RWMutex
	items map[string]cacheEntry
	ttl   time.Duration
}

func newTTLCache(ttl time.Duration) *ttlCache {
	return &ttlCache{items: make(map[string]cacheEntry), ttl: ttl}
}

func (c *ttlCache) get(key string) ([]byte, bool) {
	c.mu.RLock()
	entry, found := c.items[key]
	c.mu.RUnlock()

	if !found || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.data, true
}

func (c *ttlCache) set(key string, data []byte) {
	c.mu.Lock()
	c.items[key] = cacheEntry{data: data, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}
