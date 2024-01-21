package cache

import (
	"sync"
	"time"
)

type Cacher interface {
	Set(string, []byte, time.Duration) error
	Get(string) ([]byte, bool)
	Delete(string)
}

type Cache struct {
	data    *sync.Map
	fetcher func()
}

type entry struct {
	key    string
	value  []byte
	expire time.Time
}

func NewCache() *Cache {
	cache := &Cache{data: &sync.Map{}}
	return cache
}

func (c *Cache) Set(key string, value []byte, ttl time.Duration) error {
	entry := entry{key: key, value: value, expire: time.Now().Add(ttl)}
	c.data.Store(key, entry)
	return nil
}

func (c *Cache) Get(key string) ([]byte, bool) {
	if item, hit := c.data.Load(key); hit {
		entry := item.(entry)
		if !entry.expire.IsZero() && entry.expire.Before(time.Now()) {
			c.data.Delete(key)
			return nil, false
		}
		return entry.value, true
	}
	return nil, false
}

func (c *Cache) Delete(key string) {
	c.data.Delete(key)
}
