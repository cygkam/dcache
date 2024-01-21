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

func NewCachePool(f Fetcher) *CachePool {
	cp := &CachePool{
		peers:      hash.NewCustomConsistentHash(defaultReplicas, nil),
		localCache: cache.NewCache(),
		fetcher:    f,
	}

	return cp
}

func (cp *CachePool) SetPeers(peers ...string) {
	for _, peer := range peers {
		cp.peers.Add(peer)
	}
}

func (cp *CachePool) Get(ctx context.Context, key string) ([]byte, error) {
	value, hit := cp.localCache.Get(key)
	if hit {
		return value, nil
	}

	if peer, ok := cp.peers.Get(key); ok {
		value, err := cp.getFromPeer(ctx, peer.(string), key)
		if err == nil {
			return value, nil
		}
	}

	bytes, err := cp.fetcher.Fetch(ctx, key)
	if err != nil {
		return nil, err
	}

	cp.localCache.Set(key, bytes, cp.ttl)

	return bytes, nil
}

func (cp *CachePool) Serve(w http.ResponseWriter, r *http.Request) {
	pathParam := strings.Split(r.URL.Path, "/")

	value, err := cp.Get(r.Context(), pathParam[1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

func (cp *CachePool) getFromPeer(ctx context.Context, peer string, key string) ([]byte, error) {
	url := fmt.Sprintf("%v:%v/%v", peer, cp.port, url.QueryEscape(key))

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

	return respBody, nil
}
