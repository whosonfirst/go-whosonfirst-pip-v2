package cache

import (
	"errors"
	"fmt"
	gocache "github.com/patrickmn/go-cache"
	"github.com/whosonfirst/go-whosonfirst-log"
	"sync/atomic"
	"time"
)

type GoCache struct {
	Cache
	Logger    *log.WOFLogger
	Options   *GoCacheOptions
	cache     *gocache.Cache
	hits      int64
	misses    int64
	evictions int64
	keys      int64
}

type GoCacheOptions struct {
	CacheSize         int
	CacheTrigger      int
	DefaultExpiration time.Duration
	CleanupInterval   time.Duration
}

func (o *GoCacheOptions) String() string {
	return fmt.Sprintf("cache size %d cache trigger %d", o.CacheSize, o.CacheTrigger)
}

func DefaultGoCacheOptions() (*GoCacheOptions, error) {

	opts := GoCacheOptions{
		CacheSize:         0,
		CacheTrigger:      0,
		DefaultExpiration: 0 * time.Second,
		CleanupInterval:   0 * time.Second,
	}

	return &opts, nil
}

func NewGoCache(opts *GoCacheOptions) (Cache, error) {

	logger := log.SimpleWOFLogger("gocache")

	// mmmmmmmaybe...? how/what/where would we serialize the
	// data to...? (20170918/thisisaaronland)
	//
	// https://godoc.org/github.com/patrickmn/go-cache#NewFrom
	// https://godoc.org/github.com/patrickmn/go-cache#Cache.Items
	// https://godoc.org/github.com/patrickmn/go-cache#Item

	c := gocache.New(opts.DefaultExpiration, opts.CleanupInterval)

	lc := GoCache{
		Logger:    logger,
		Options:   opts,
		cache:     c,
		hits:      int64(0),
		misses:    int64(0),
		evictions: int64(0),
		keys:      0,
	}

	return &lc, nil
}

func (c *GoCache) Close() error {
	return nil
}

func (c *GoCache) Get(key string) (CacheItem, error) {

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

func (c *GoCache) Set(key string, item CacheItem) error {

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

	c.cache.Set(key, item, gocache.DefaultExpiration)
	atomic.AddInt64(&c.keys, 1)

	return nil
}

func (c *GoCache) Size() int64 {
	return atomic.LoadInt64(&c.keys)
}

func (c *GoCache) Hits() int64 {
	return atomic.LoadInt64(&c.hits)
}

func (c *GoCache) Misses() int64 {
	return atomic.LoadInt64(&c.misses)
}

func (c *GoCache) Evictions() int64 {
	return atomic.LoadInt64(&c.evictions)
}
