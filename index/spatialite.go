package index

// https://gist.github.com/simonw/91a1157d1f45ab305c6f48c4ca344de8
// http://www.gaia-gis.it/gaia-sins/spatialite-sql-4.3.0.html

// ./bin/wof-sqlite-index-features -driver spatialite -dsn test-pip.db -geojson -geometries -live-hard-die-fast -timings -mode repo /usr/local/data/whosonfirst-data-constituency-us

import (
	"errors"
	"fmt"
	"github.com/skelterjohn/geom"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-pip"
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	"github.com/whosonfirst/go-whosonfirst-spr"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"io"
	"os"
	"sync"
)

type SpatialiteIndex struct {
	Index
	Logger   *log.WOFLogger
	database *database.SQLiteDatabase
	cache    cache.Cache
	mu       *sync.RWMutex
	throttle chan bool
}

type SpatialiteResults struct {
	spr.StandardPlacesResults `json:",omitempty"`
	Places                    []spr.StandardPlacesResult `json:"places"`
}

func (r *SpatialiteResults) Results() []spr.StandardPlacesResult {
	return r.Places
}

func NewSpatialiteIndex(db *database.SQLiteDatabase, c cache.Cache) (Index, error) {

	logger := log.SimpleWOFLogger("index")

	stdout := io.Writer(os.Stdout)
	logger.AddLogger(stdout, "info")

	_, err := tables.NewGeometriesTableWithDatabase(db)

	if err != nil {
		return nil, err
	}

	mu := new(sync.RWMutex)

	// PLEASE TO ADD CONNECTION POOLS TO
	// SQLITE THINGY (20180221/thisisaaronland)

	maxconns := 32
	throttle := make(chan bool, maxconns)

	for i := 0; i < maxconns; i++ {
		throttle <- true
	}

	i := SpatialiteIndex{
		database: db,
		cache:    c,
		Logger:   logger,
		mu:       mu,
		throttle: throttle,
	}

	return &i, nil
}

func (i *SpatialiteIndex) Cache() cache.Cache {
	return i.cache
}

func (i *SpatialiteIndex) IndexFeature(f geojson.Feature) error {

	// SEE ABOVE

	<-i.throttle

	defer func() {
		i.throttle <- true
	}()

	i.mu.Lock()
	defer i.mu.Unlock()

	db := i.database

	t, err := tables.NewGeometriesTable()

	if err != nil {
		return err
	}

	fc, err := cache.NewFeatureCache(f)

	if err != nil {
		return err
	}

	str_id := f.Id()

	err = i.cache.Set(str_id, fc)

	if err != nil {
		return err
	}

	return t.IndexRecord(db, f)
}

func (i *SpatialiteIndex) GetIntersectsByCoord(coord geom.Coord, f filter.Filter) (spr.StandardPlacesResults, error) {

	db := i.database

	conn, err := db.Conn()

	if err != nil {
		return nil, err
	}

	lat := coord.Y
	lon := coord.X

	places := make([]spr.StandardPlacesResult, 0)

	// for reasons I don't understand this returns empty - I am guessing it has something
	// to do with internal escaping... (20180220/thisisaaronland)
	// q := `SELECT id FROM geometries WHERE ST_Within(GeomFromText('POINT(? ?)'), geom) AND rowid IN (SELECT pkid FROM idx_geometries_geom WHERE xmin < ? AND xmax > ? AND ymin < ? AND ymax > ?)`
	// rows, err := conn.Query(q, lon, lat, lon, lon, lat, lat)

	q := fmt.Sprintf(`SELECT id FROM geometries WHERE ST_Within(GeomFromText('POINT(%0.6f %0.6f)'), geom)
		          AND rowid IN (
			    SELECT pkid FROM idx_geometries_geom WHERE xmin < %0.6f AND xmax > %0.6f AND ymin < %0.6f AND ymax > %0.6f
                          )`, lon, lat, lon, lon, lat, lat)

	i.Logger.Info("QUERY %s", q)

	rows, err := conn.Query(q)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {

		var str_id string
		err = rows.Scan(&str_id)

		if err != nil {
			return nil, err
		}

		fc, err := i.cache.Get(str_id)

		if err != nil {
			i.Logger.Info("SAD")
			return nil, err
		}

		s := fc.SPR()

		err = filter.FilterSPR(f, s)

		if err != nil {
			i.Logger.Debug("SKIP %s because filter error %s", str_id, err)
			continue
		}

		places = append(places, fc.SPR())
	}

	err = rows.Err()

	if err != nil {
		return nil, err
	}

	r := SpatialiteResults{
		Places: places,
	}

	i.Logger.Info("RETURN")
	return &r, nil
}

func (i *SpatialiteIndex) GetCandidatesByCoord(coord geom.Coord) (*pip.GeoJSONFeatureCollection, error) {

	db := i.database

	conn, err := db.Conn()

	if err != nil {
		return nil, err
	}

	lat := coord.Y
	lon := coord.X

	// ORDER BY... ?

	q := `SELECT id, AsGeoJSON(ST_Envelope(geom)) AS geom FROM geometries WHERE ST_Within(GeomFromText('POINT(? ?)'), ST_Envelope(geom))`

	rows, err := conn.Query(q, lon, lat)

	if err != nil {
		return nil, err
	}

	for rows.Next() {

		var str_id string
		var str_geom string

		err = rows.Scan(&str_id, str_geom)

		if err != nil {
			return nil, err
		}

		i.Logger.Status("PLEASE WRITE ME %s %s", str_id, str_geom)
	}

	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return nil, errors.New("PLEASE WRITE ME")
}

func (i *SpatialiteIndex) GetIntersectsByPath(path geom.Path, f filter.Filter) ([]spr.StandardPlacesResults, error) {
	return nil, errors.New("PLEASE WRITE ME")
}
