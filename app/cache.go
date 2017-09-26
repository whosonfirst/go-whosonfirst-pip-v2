package app

import (
	"context"
	"errors"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/utils"
	"io"
	"os"
	"sync/atomic"
)

type ApplicationCacheOptions struct {
	FailoverCache       bool
	FailoverCacheEngine string
	GoCache             bool
	IndexMode           string
	IndexPaths          []string
	LRUCache            bool
	LRUCacheSize        int
	LRUCacheTriggerSize int
	SourceCache         bool
	SourceCacheRoot     string
}

func DefaultApplicationCacheOptions() (ApplicationCacheOptions, error) {

	opts := ApplicationCacheOptions{
		FailoverCache:       false,
		FailoverCacheEngine: "",
		IndexMode:           "",
		IndexPaths:          make([]string, 0),
		GoCache:             false,
		LRUCache:            false,
		LRUCacheSize:        0,
		LRUCacheTriggerSize: 0,
		SourceCache:         false,
		SourceCacheRoot:     "",
	}

	return opts, nil
}

func ApplicationCache(opts ApplicationCacheOptions) (cache.Cache, error) {

	var failover_cache cache.Cache
	var source_cache cache.Cache
	var lru_cache cache.Cache
	var pip_cache cache.Cache
	var go_cache cache.Cache

	if opts.FailoverCache {

		opts.SourceCache = true

		switch opts.FailoverCacheEngine {
		case "lru":
			opts.LRUCache = true
		case "gocache":
			opts.GoCache = true
		default:
			return nil, errors.New("Invalid failover cache engine")
		}

	}

	if opts.GoCache {

		opts, err := cache.DefaultGoCacheOptions()

		if err != nil {
			return nil, err
		}

		c, err := cache.NewGoCache(opts)

		if err != nil {
			return nil, err
		}

		go_cache = c
		pip_cache = go_cache
	}

	if opts.SourceCache {

		_, err := os.Stat(opts.SourceCacheRoot)

		if os.IsNotExist(err) {
			return nil, err
		}

		c, err := cache.NewSourceCache(opts.SourceCacheRoot)

		if err != nil {
			return nil, err
		}

		source_cache = c
		pip_cache = source_cache
	}

	if opts.LRUCache {

		sz := int32(opts.LRUCacheSize)

		if sz == 0 {

			cb := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

				ok, err := utils.IsValidRecord(fh, ctx)

				if err != nil {
					return err
				}

				if !ok {
					return nil
				}

				atomic.AddInt32(&sz, 1)
				return nil
			}

			idx, err := index.NewIndexer(opts.IndexMode, cb)

			if err != nil {
				return nil, err
			}

			err = idx.IndexPaths(opts.IndexPaths)

			if err != nil {
				return nil, err
			}
		}

		o, err := cache.DefaultLRUCacheOptions()

		if err != nil {
			return nil, err
		}

		o.CacheSize = int(sz)
		o.CacheTrigger = opts.LRUCacheTriggerSize

		c, err := cache.NewLRUCache(o)

		if err != nil {
			return nil, err
		}

		lru_cache = c
		pip_cache = lru_cache
	}

	if opts.FailoverCache {

		c, err := cache.NewFailoverCache(lru_cache, source_cache)

		if err != nil {
			return nil, err
		}

		failover_cache = c
		pip_cache = failover_cache
	}

	return pip_cache, nil
}
