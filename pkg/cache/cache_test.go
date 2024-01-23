package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetShouldEndWithSuccess(t *testing.T) {
	c := NewCache()
	key := "9663d0d0-eb9d-429a-8c87-6c8b6645bee7"
	data := []byte("Ut fringilla lectus nec")

	c.Set(key, data, 0)
	cachedEntry, hit := c.data.Load(key)
	e := cachedEntry.(entry)

	assert.True(t, hit)
	assert.Equal(t, key, e.key)
	assert.Equal(t, data, e.value)
}

func TestGetShouldGiveCacheMiss(t *testing.T) {
	c := NewCache()
	key := "66b68226-22b6-4da2-a072-6fc3a684618e"

	cachedData, hit := c.Get(key)

	assert.False(t, hit)
	assert.Equal(t, cachedData, []byte(nil))
}

func TestGetShouldGiveCacheExpiration(t *testing.T) {
	c := NewCache()
	ttl := time.Second * 1
	delay := time.Second * 3
	key := "2d8b59ca-fce0-4645-8de3-ec88f899656f"
	data := []byte("Cras ac lectus ut lectus")
	item := entry{key: key, value: data, expire: time.Now().Add(ttl)}

	c.data.Store(key, item)
	time.Sleep(delay)
	cachedData, hit := c.Get(key)

	assert.False(t, hit)
	assert.Equal(t, cachedData, []byte(nil))
}

func TestGetShouldEndWithSuccess(t *testing.T) {
	c := NewCache()
	ttl := time.Second * 3
	delay := time.Second * 1
	key := "7fbd1d90-5ff7-4f1d-a3b4-e79b5ccf3e12"
	data := []byte("Maecenas tincidunt nulla")
	item := entry{key: key, value: data, expire: time.Now().Add(ttl)}

	c.data.Store(key, item)
	time.Sleep(delay)
	cachedData, hit := c.Get(key)

	assert.True(t, hit)
	assert.Equal(t, cachedData, data)
}

func TestDeleteShouldEndWithSuccess(t *testing.T) {
	c := NewCache()
	ttl := time.Second * 30
	key := "7fbd1d90-5ff7-4f1d-a3b4-e79b5ccf3e12"
	data := []byte("Maecenas tincidunt nulla")
	item := entry{key: key, value: data, expire: time.Now().Add(ttl)}

	c.data.Store(key, item)
	c.Delete(key)
	cachedEntry, hit := c.data.Load(key)

	assert.False(t, hit)
	assert.Equal(t, cachedEntry, nil)
}
