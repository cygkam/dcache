package cachepool

import (
	"context"
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
	defaultPort     = "9929"
)

type (
	Fetcher interface {
		Fetch(ctx context.Context, key string) ([]byte, error)
	}

	CachePoolCfg struct {
		Fetcher Fetcher
		Port    string
		Ttl     time.Duration
	}

	CachePool struct {
		fetcher      Fetcher
		fetchEnabled bool
		localCache   *cache.Cache
		peers        *hash.ConsistentHash
		port         string
		ttl          time.Duration
	}
)

func NewCachePool(cfg *CachePoolCfg) *CachePool {
	var (
		port         string
		fetcher      Fetcher
		fetchEnabled bool
	)

	if cfg.Port == "" {
		port = defaultPort
	} else {
		port = cfg.Port
	}

	if cfg.Fetcher == nil {
		fetchEnabled = false
	} else {
		fetcher = cfg.Fetcher
		fetchEnabled = true
	}

	cp := &CachePool{
		peers:        hash.NewCustomConsistentHash(defaultReplicas, nil),
		localCache:   cache.NewCache(),
		fetcher:      fetcher,
		fetchEnabled: fetchEnabled,
		port:         port,
		ttl:          cfg.Ttl,
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
	}

	return nil, false
}

func (cp *CachePool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pathParam := strings.TrimPrefix(r.URL.Path, "/")

	bytes, err := cp.getFromLocalCache(r.Context(), pathParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Write(bytes)
}

func (cp *CachePool) getFromLocalCache(ctx context.Context, key string) ([]byte, error) {
	value, hit := cp.localCache.Get(key)
	if hit {
		return value, nil
	}

	if cp.fetchEnabled {
		value, err := cp.fetcher.Fetch(ctx, key)
		if err != nil {
			return value, err
		}
		cp.localCache.Set(key, value, cp.ttl)
		return value, nil
	}

	return nil, nil
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

	return respBody, nil
}
