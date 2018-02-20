package cache

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-pip"
	"github.com/whosonfirst/go-whosonfirst-spr"
	// "github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"github.com/whosonfirst/go-whosonfirst-sqlite/utils"
	"sync/atomic"
)

type SpatialiteCache struct {
	Cache
	Logger    *log.WOFLogger
	Options   *SpatialiteCacheOptions
	database  *database.SQLiteDatabase
	hits      int64
	misses    int64
	evictions int64
}

type SpatialiteCacheOptions struct {
	Set bool // PLEASE RENAME ME
}

func DefaultSpatialiteCacheOptions() (*SpatialiteCacheOptions, error) {

	opts := SpatialiteCacheOptions{
		Set: true,
	}

	return &opts, nil
}

func NewSpatialiteCache(db *database.SQLiteDatabase, opts *SpatialiteCacheOptions) (Cache, error) {

	logger := log.SimpleWOFLogger("spatialite")

	ok_geojson, err := utils.HasTable(db, "geojson")

	if err != nil {
		return nil, err
	}

	if !ok_geojson {
		return nil, errors.New("Missing 'geojson' table")
	}

	lc := SpatialiteCache{
		Logger:    logger,
		Options:   opts,
		database:  db,
		hits:      int64(0),
		misses:    int64(0),
		evictions: int64(0),
	}

	return &lc, nil
}

func (c *SpatialiteCache) Get(key string) (CacheItem, error) {

	c.Logger.Info("GET %s", key)

	db := c.database

	conn, err := db.Conn()

	if err != nil {
		return nil, err
	}

	q := "SELECT body FROM geojson WHERE id = ?"
	row := conn.QueryRow(q, key)

	var body string
	err = row.Scan(&body)

	if err != nil {

		if err == sql.ErrNoRows {
			atomic.AddInt64(&c.misses, 1)
			return nil, errors.New("CACHE MISS")
		}

		return nil, err
	}

	f, err := feature.LoadFeature([]byte(body))

	if err != nil {
		return nil, err
	}

	fc, err := NewFeatureCache(f)

	if err != nil {
		return nil, err
	}

	atomic.AddInt64(&c.hits, 1)
	return fc, nil
}

func (c *SpatialiteCache) Set(key string, item CacheItem) error {

	if !c.Options.Set {
		return nil
	}

	// PLEASE RECONCILE THIS CODE WITH
	// go-whosonfirst-sqlite-features/tables/geojson.go

	db := c.database

	conn, err := db.Conn()

	if err != nil {
		return err
	}

	s := item.SPR()
	g := item.Geometry()

	str_id := s.Id()
	lastmod := s.LastModified()

	type Feature struct {
		Geometry   pip.GeoJSONGeometry      `json:"geometry"`
		Properties spr.StandardPlacesResult `json:"properties"`
	}

	f := Feature{
		Geometry:   g,
		Properties: s,
	}

	body, err := json.Marshal(f)

	if err != nil {
		return nil
	}

	tx, err := conn.Begin()

	if err != nil {
		return err
	}

	q := "INSERT OR REPLACE INTO geojson (id, body, lastmodified) VALUES (?, ?, ?)"

	stmt, err := tx.Prepare(q)

	if err != nil {
		return err
	}

	defer stmt.Close()

	str_body := string(body)

	_, err = stmt.Exec(str_id, str_body, lastmod)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (c *SpatialiteCache) Size() int64 {

	db := c.database

	conn, err := db.Conn()

	if err != nil {
		return -1
	}

	q := "SELECT COUNT(id) FROM geojson"
	row := conn.QueryRow(q)

	var count int64
	err = row.Scan(&count)

	if err != nil {

		return -1
	}

	return count
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
