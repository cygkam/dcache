
# DCache library for distributed caching across many instances

This library was implemented as part of a recruitment task. It was intended to demonstrate the problem that occurs when sharing shared cache memory between multiple instances of an application.

## Features
- Local cache operations: Set, Get, and Delete
- Setting TTL - the maximum lifespan of a data in a cache before it is discarded
- Lazy loading - loads data into the cache only when necessary using auto fetching mechanism
- Distribution across cache pool peers using consistent hashing

## Usage/Examples

### Local cache usage
```golang
package main

import "github.com/cygkam/dcache/pkg/cache"

func main() {

  c := cache.NewCache()
  key := "9663d0d0-eb9d-429a-8c87-6c8b6645bee7"
  data := []byte("Ut fringilla lectus nec")
  ttl := time.Second * 30

  c.Set(key, data, ttl)

  cachedData, hit := c.Get(key)
  if hit {
    fmt.Println(string(cachedData))
    // Ut fringilla lectus nec
  }
  
  c.Delete(key)
}
```
### Distributed cache usage
```golang
package main

import "github.com/cygkam/dcache/pkg/cachepool"

func main() {
  cfg := &cachepool.CachePoolCfg{
		Port:    "30345",
		Ttl:     time.Second * 30
  }

  cp := cachepool.NewCachePool(cfg)
  cp.SetPeers("10.11.0.1:9909","10.11.0.2:9909")
  go cp.StartHTTPPoolServer()

  cachedData, hit := c.Get(key)
  if hit {
    fmt.Println(string(cachedData))
    // Who knows what peers have in their local cache
  }
}
```

## Links to articles that helped
- https://blog.stackademic.com/distributed-data-done-right-a-golang-journey-with-consistent-hashing-49125707165b
- https://dev.to/kevwan/consistent-hash-in-go-3aob
- http://highscalability.com/blog/2023/2/22/consistent-hashing-algorithm.html
