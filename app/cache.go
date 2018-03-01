package app

import (
	"errors"
	"flag"
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/flags"
)

func NewApplicationCache(fl *flag.FlagSet) (cache.Cache, error) {

     pip_cache, err := flags.Lookup(fl, "pip-cache")

     if err != nil {
     	return nil, err
     }

	switch pip_cache {

	case "gocache":

		opts, err := cache.DefaultGoCacheOptions()

		if err != nil {
			return nil, err
		}

		return cache.NewGoCache(opts)

	case "fs":

		path, err := flags.Lookup(fl, "fs-path")

		if err != nil {
			return nil, err
		}

		appcache, appcache_err = cache.NewFSCache(path)

	case "sqlite":
		return cache.NewSQLiteCache(db)
	case "spatialite":
		return cache.NewSQLiteCache(db)
	default:
		return nil, errors.New("Invalid cache layer")
	}

}
