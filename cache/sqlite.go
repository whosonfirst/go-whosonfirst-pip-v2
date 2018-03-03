package cache

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-pip"
	"github.com/whosonfirst/go-whosonfirst-spr"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"sync/atomic"
)

type SQLiteCache struct {
	Cache
	Logger    *log.WOFLogger
	database  *database.SQLiteDatabase
	hits      int64
	misses    int64
	evictions int64
}

func NewSQLiteCache(db *database.SQLiteDatabase) (Cache, error) {

	logger := log.SimpleWOFLogger("sqlite")

	_, err := tables.NewGeoJSONTableWithDatabase(db)

	if err != nil {
		return nil, err
	}

	lc := SQLiteCache{
		Logger:    logger,
		database:  db,
		hits:      int64(0),
		misses:    int64(0),
		evictions: int64(0),
	}

	return &lc, nil
}

func (c *SQLiteCache) Close() error {
	return c.database.Close()
}

func (c *SQLiteCache) Get(key string) (CacheItem, error) {

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

func (c *SQLiteCache) Set(key string, item CacheItem) error {

	// PLEASE RECONCILE THIS CODE WITH
	// go-whosonfirst-sqlite-features/tables/geojson.go

	// what that means in practical terms is we need to write
	// something that takes a cache item and implements all of
	// the geojson.Feature interface and write now we're just
	// making something that sort of looks like it...
	// (20180224/thisisaaronland)

	// this is more complicated that it seems (or should be) so
	// see notes in cmd/wof-pip-server.go in the 'if *enable_extras'
	// section (20180228/thisisaaronland)

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

func (c *SQLiteCache) Size() int64 {

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

func (c *SQLiteCache) Hits() int64 {
	return atomic.LoadInt64(&c.hits)
}

func (c *SQLiteCache) Misses() int64 {
	return atomic.LoadInt64(&c.misses)
}

func (c *SQLiteCache) Evictions() int64 {
	return atomic.LoadInt64(&c.evictions)
}
