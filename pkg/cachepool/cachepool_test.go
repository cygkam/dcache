package cachepool

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MockFetcher struct {
	data []byte
	err  error
}

func (mf *MockFetcher) Fetch(ctx context.Context, key string) ([]byte, error) {
	return mf.data, mf.err
}

func TestNewCachePoolShouldCreateDefaultCachePool(t *testing.T) {
	cfg := &CachePoolCfg{}

	cp := NewCachePool(cfg)

	assert.Equal(t, cp.ttl, time.Duration(0))
	assert.Equal(t, cp.port, "9929")
	assert.Equal(t, cp.fetchEnabled, false)
	assert.Equal(t, cp.fetcher, nil)
}

func TestNewCachePoolShouldCreateCustomCachePool(t *testing.T) {
	mf := &MockFetcher{}
	cfg := &CachePoolCfg{
		Port:    "30345",
		Ttl:     time.Second * 30,
		Fetcher: mf,
	}

	cp := NewCachePool(cfg)

	assert.Equal(t, cp.ttl, time.Second*30)
	assert.Equal(t, cp.port, "30345")
	assert.Equal(t, cp.fetchEnabled, true)
	assert.Equal(t, cp.fetcher, mf)
}

func TestGetShouldRetrieveValueFromLocalCache(t *testing.T) {
	key := "2d8b59ca-fce0-4645-8de3-ec88f899656f"
	data := []byte("Cras ac lectus ut lectus")
	cfg := &CachePoolCfg{
		Fetcher: &MockFetcher{},
		Ttl:     time.Second * 30,
	}
	cp := NewCachePool(cfg)

	cp.localCache.Set(key, data, cp.ttl)
	ctx := context.Background()
	cachedData, hit := cp.Get(ctx, key)

	assert.True(t, hit)
	assert.Equal(t, data, cachedData)
}

func TestServeHTTPShouldReturnValueFromLocalCache(t *testing.T) {
	key := "2d8b59ca-fce0-4645-8de3-ec88f899656f"
	data := []byte("Cras ac lectus ut lectus")
	cfg := &CachePoolCfg{
		Fetcher: &MockFetcher{},
		Ttl:     time.Second * 30,
	}
	cp := NewCachePool(cfg)

	cp.localCache.Set(key, data, cp.ttl)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s", key), nil)
	cp.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, data, recorder.Body.Bytes())
}

func TestServeHTTPShouldGiveCacheMiss(t *testing.T) {
	key := "2d8b59ca-fce0-4645-8de3-ec88f899656f"
	cfg := &CachePoolCfg{
		Ttl: time.Second * 30,
	}
	cp := NewCachePool(cfg)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s", key), nil)
	cp.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, []byte(nil), recorder.Body.Bytes())
}

func TestServeHTTPShouldReturnValueFetchedFromSourceAndPopulateLocally(t *testing.T) {
	key := "2d8b59ca-fce0-4645-8de3-ec88f899656f"
	data := []byte("Cras ac lectus ut lectus")
	cfg := &CachePoolCfg{
		Fetcher: &MockFetcher{data: data},
		Ttl:     time.Second * 30,
	}
	cp := NewCachePool(cfg)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s", key), nil)
	cp.ServeHTTP(recorder, req)

	dataEntry, hit := cp.localCache.Get(key)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, data, recorder.Body.Bytes())
	assert.True(t, hit)
	assert.Equal(t, data, dataEntry)
}
