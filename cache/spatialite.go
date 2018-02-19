package cache

import (
	"errors"
	_ "fmt"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"github.com/whosonfirst/go-whosonfirst-sqlite/utils"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
	"sync/atomic"
)

type SpatialiteCache struct {
	Cache
	Logger    *log.WOFLogger
	Options   *SpatialiteCacheOptions
	hits      int64
	misses    int64
	evictions int64
	keys      int64
	size      int64
}

type SpatialiteCacheOptions struct {
     DB *database.SQLiteDatabase
}

func DefaultSpatialiteCacheOptions() (*SpatialiteCacheOptions, error) {

	db, err := database.NewDBWithDriver("spatialite", ":memory:")

	if err != nil {
		return nil, err
	}

	_, err = tables.NewSPRTableWithDatabase(db)

	if err != nil {
		return nil, err
	}

	_, err = tables.NewGeometriesTableWithDatabase(db)

	if err != nil {
		return nil, err
	}

	opts := SpatialiteCacheOptions{
		DB: db,
	}

	return &opts, nil
}

func NewSpatialiteCache(opts *SpatialiteCacheOptions) (Cache, error) {

	logger := log.SimpleWOFLogger("spatialite")

	ok_spr, err := utils.HasTable(opts.DB, "spr")

	if err != nil {
		return nil, err
	}

	if !ok_spr {
		return nil, errors.New("Missing 'spr' table")
	}

	ok_geom, err := utils.HasTable(opts.DB, "geometries")

	if err != nil {
		return nil, err
	}

	if !ok_geom {
		return nil, errors.New("Missing 'geometries' table")
	}
	
	lc := SpatialiteCache{
		Logger:    logger,
		Options:   opts,
		hits:      int64(0),
		misses:    int64(0),
		evictions: int64(0),
		keys:      0,
		size:      0,
	}

	return &lc, nil
}

func (c *SpatialiteCache) Get(key string) (CacheItem, error) {

	// to do: timings that don't slow everything down the way
	// go-whosonfirst-timer does now (20170915/thisisaaronland)

	c.Logger.Info("GET %s", key)

	return nil, errors.New("PLEASE FINISH ME")

	/*
	db := c.Options.db

	conn, err := db.Conn()

	if err != nil {
		return err
	}

	sql := fmt.Sprintf("SELECT * FROM spr WHERE id = ?"


	if !ok {
		atomic.AddInt64(&c.misses, 1)
		return nil, errors.New("CACHE MISS")
	}

	atomic.AddInt64(&c.hits, 1)

	fc := cache.(CacheItem)
	return fc, nil
	*/
}

func (c *SpatialiteCache) Set(key string, item CacheItem) error {

     return nil
}

func (c *SpatialiteCache) Size() int64 {
	return c.size
	// return atomic.LoadInt64(&c.keys)
}

func (c *SpatialiteCache) Hits() int64 {
	return atomic.LoadInt64(&c.hits)
}

func (c *SpatialiteCache) Misses() int64 {
	return atomic.LoadInt64(&c.misses)
}

func (c *SpatialiteCache) Evictions() int64 {
	return atomic.LoadInt64(&c.evictions)
}
