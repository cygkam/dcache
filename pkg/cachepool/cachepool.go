package cachepool

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cygkam/dcache/pkg/cache"
	"github.com/zeromicro/go-zero/core/hash"
)

const (
	defaultReplicas = 40
)

type HashFunc func([]byte) uint64

type Fetcher interface {
	Fetch(ctx context.Context, key string) ([]byte, error)
}

type Fetch func(ctx context.Context, key string) ([]byte, error)

type CachePool struct {
	peers      *hash.ConsistentHash
	localCache *cache.Cache
	fetcher    Fetcher
	port       string
	ttl        time.Duration
}

func NewCachePool(f Fetcher, ttl time.Duration, port string) *CachePool {
	cp := &CachePool{
		peers:      hash.NewCustomConsistentHash(defaultReplicas, nil),
		localCache: cache.NewCache(),
		fetcher:    f,
		port:       port,
		ttl:        ttl,
	}

	return cp
}

func (cp *CachePool) SetPeers(peers ...string) {
	for _, peer := range peers {
		cp.peers.Add(peer)
	}
}

func (cp *CachePool) Get(ctx context.Context, key string) ([]byte, bool) {
	if value, hit := cp.localCache.Get(key); hit {
		return value, true
	}

	if peer, ok := cp.peers.Get(key); ok {
		value, err := cp.getFromPeer(ctx, peer.(string), key)
		if err == nil {
			cp.localCache.Set(key, value, cp.ttl)
			return value, true
		}
		fmt.Printf("get error: %s\n", err.Error())
	}

	fmt.Println("No peer found")

	return nil, false
}

func (cp *CachePool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pathParam := strings.TrimPrefix(r.URL.Path, "/")

	value, err := cp.getFromLocalCache(r.Context(), pathParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	bytes, err := json.Marshal(value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

func (cp *CachePool) getFromLocalCache(ctx context.Context, key string) ([]byte, error) {
	fmt.Println("Get from local cache - peer")
	value, hit := cp.localCache.Get(key)
	if hit {
		return value, nil
	}
	value, err := cp.fetcher.Fetch(ctx, key)
	if err != nil {
		return value, err
	}

	cp.localCache.Set(key, value, cp.ttl)

	return value, nil
}

func (cp *CachePool) getFromPeer(ctx context.Context, peer string, key string) ([]byte, error) {
	url := fmt.Sprintf("http://%v/%v", peer, url.QueryEscape(key))

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	fmt.Sprintln(respBody)
	return respBody, nil
}
