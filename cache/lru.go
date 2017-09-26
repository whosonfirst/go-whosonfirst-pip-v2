package cache

import (
	"errors"
	"fmt"
	"github.com/hashicorp/golang-lru"
	"github.com/whosonfirst/go-whosonfirst-log"
	"sync/atomic"
)

type LRUCache struct {
	Cache
	Logger    *log.WOFLogger
	Options   *LRUCacheOptions
	cache     *lru.Cache
	hits      int64
	misses    int64
	evictions int64
	keys      int64
	size      int64
}

type LRUCacheOptions struct {
	CacheSize    int
	CacheTrigger int
}

func (o *LRUCacheOptions) String() string {
	return fmt.Sprintf("cache size %d cache trigger %d", o.CacheSize, o.CacheTrigger)
}

func DefaultLRUCacheOptions() (*LRUCacheOptions, error) {

	opts := LRUCacheOptions{
		CacheSize:    0,
		CacheTrigger: 0,
	}

	return &opts, nil
}

func NewLRUCache(opts *LRUCacheOptions) (Cache, error) {

	logger := log.SimpleWOFLogger("lru")
	// logger.Status("%s", opts)

	c, err := lru.New(opts.CacheSize)

	if err != nil {
		return nil, err
	}

	lc := LRUCache{
		Logger:    logger,
		Options:   opts,
		cache:     c,
		hits:      int64(0),
		misses:    int64(0),
		evictions: int64(0),
		keys:      0,
		size:      int64(opts.CacheSize),
	}

	return &lc, nil
}

func (c *LRUCache) Get(key string) (CacheItem, error) {

	// to do: timings that don't slow everything down the way
	// go-whosonfirst-timer does now (20170915/thisisaaronland)

	c.Logger.Info("GET %s", key)

	cache, ok := c.cache.Get(key)

	if !ok {
		atomic.AddInt64(&c.misses, 1)
		return nil, errors.New("CACHE MISS")
	}

	atomic.AddInt64(&c.hits, 1)

	fc := cache.(CacheItem)
	return fc, nil
}

func (c *LRUCache) Set(key string, item CacheItem) error {

	var points int

	if c.Options.CacheTrigger > 0 {

		points = 0

		for _, p := range item.Polygons() {

			ext := p.ExteriorRing()
			points += ext.Length()

			if points >= c.Options.CacheTrigger {
				break
			}
		}

		if points < c.Options.CacheTrigger {
			c.Logger.Debug("SKIP %s INSUFFICIENT POINTS %d", key, points)
			return nil
		}
	}

	// c.Logger.Debug("SET %s %d points", key, points)

	evicted := c.cache.Add(key, item)
	atomic.AddInt64(&c.keys, 1)

	if evicted {
		atomic.AddInt64(&c.evictions, 1)
		atomic.AddInt64(&c.keys, -1)
		c.Logger.Warning("EVICTION caused by %s", key)
	}

	return nil
}

func (c *LRUCache) Size() int64 {
	return c.size
	// return atomic.LoadInt64(&c.keys)
}

func (c *LRUCache) Hits() int64 {
	return atomic.LoadInt64(&c.hits)
}

func (c *LRUCache) Misses() int64 {
	return atomic.LoadInt64(&c.misses)
}

func (c *LRUCache) Evictions() int64 {
	return atomic.LoadInt64(&c.evictions)
}
