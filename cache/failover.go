package cache

import (
	"github.com/whosonfirst/go-whosonfirst-log"
	"sync/atomic"
)

type FailoverCache struct {
	Cache
	Logger         *log.WOFLogger
	primary_cache  Cache
	failover_cache Cache
	hits           int64
	misses         int64
	evictions      int64
	keys           int64
}

func NewFailoverCache(primary Cache, failover Cache) (Cache, error) {

	logger := log.SimpleWOFLogger("failover")

	c := FailoverCache{
		Logger:         logger,
		primary_cache:  primary,
		failover_cache: failover,
		hits:           0,
		misses:         0,
		evictions:      0,
		keys:           0,
	}

	return &c, nil
}

func (c *FailoverCache) Get(key string) (CacheItem, error) {

	// to do: timings that don't slow everything down the way
	// go-whosonfirst-timer does now (20170915/thisisaaronland)

	item, err := c.primary_cache.Get(key)

	if err == nil {
		c.Logger.Debug("PRIMARY HIT %s", key)
		atomic.AddInt64(&c.hits, 1)
		return item, nil
	}

	c.Logger.Warning("PRIMARY MISS %s", key)
	atomic.AddInt64(&c.misses, 1)

	item, err = c.failover_cache.Get(key)

	if err == nil {
		c.Logger.Debug("SECONDARY HIT %s", key)
		atomic.AddInt64(&c.hits, 1)
	} else {
		c.Logger.Warning("SECONDARY MISS %s", key)
		atomic.AddInt64(&c.misses, 1)
	}

	return item, err
}

func (c *FailoverCache) Set(key string, i CacheItem) error {

	c.Logger.Debug("SET %s", key)

	err := c.primary_cache.Set(key, i)

	if err != nil {
		return err
	}

	return c.failover_cache.Set(key, i)
}

func (c *FailoverCache) Size() int64 {
	return c.primary_cache.Size() + c.failover_cache.Size()
}

func (c *FailoverCache) Hits() int64 {
	return atomic.LoadInt64(&c.hits)
}

func (c *FailoverCache) Misses() int64 {
	return atomic.LoadInt64(&c.misses)
}

func (c *FailoverCache) Evictions() int64 {
	return c.primary_cache.Evictions() + c.failover_cache.Evictions()
}
