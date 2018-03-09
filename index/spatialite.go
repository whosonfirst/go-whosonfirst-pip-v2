package index

// https://gist.github.com/simonw/91a1157d1f45ab305c6f48c4ca344de8
// http://www.gaia-gis.it/gaia-sins/spatialite-sql-4.3.0.html

import (
	"errors"
	"fmt"
	"github.com/skelterjohn/geom"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-pip"
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	"github.com/whosonfirst/go-whosonfirst-spr"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"strings"
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

	_, err := tables.NewGeometriesTableWithDatabase(db)

	if err != nil {
		return nil, err
	}

	mu := new(sync.RWMutex)

	// PLEASE TO ADD CONNECTION POOLS TO
	// SQLITE THINGY (20180221/thisisaaronland)

	maxconns := 64
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

func (i *SpatialiteIndex) Close() error {
	return i.database.Close()
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
			return nil, err
		}

		s := fc.SPR()

		err = filter.FilterSPR(f, s)

		if err != nil {
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

	q := fmt.Sprintf(`SELECT id, AsGeoJSON(ST_Envelope(geom)) AS geom FROM geometries WHERE ST_Within(GeomFromText('POINT(%0.6f %0.6f)'), ST_Envelope(geom))`, lon, lat)

	rows, err := conn.Query(q)

	if err != nil {
		return nil, err
	}

	features := make([]pip.GeoJSONFeature, 0)

	for rows.Next() {

		var str_id string
		var str_geom string

		err = rows.Scan(&str_id, &str_geom)

		if err != nil {
			return nil, err
		}

		props := map[string]interface{}{
			"id": str_id,
		}

		// this should be easier than this... but it's not
		// (20180225/thisisaaronland)

		coords := gjson.GetBytes([]byte(str_geom), "coordinates")

		if !coords.Exists() {
			return nil, errors.New("Invalid coordinates")
		}

		ring := make([]pip.GeoJSONPoint, 0)

		for _, r := range coords.Array() {

			for _, p := range r.Array() {

				lonlat := p.Array()

				lon := lonlat[0].Float()
				lat := lonlat[1].Float()

				pt := pip.GeoJSONPoint{lon, lat}
				ring = append(ring, pt)
			}
		}

		poly := pip.GeoJSONPolygon{ring}
		multi := pip.GeoJSONMultiPolygon{poly}

		geom := pip.GeoJSONGeometry{
			Type:        "MultiPolygon",
			Coordinates: multi,
		}

		feature := pip.GeoJSONFeature{
			Type:       "Feature",
			Properties: props,
			Geometry:   geom,
		}

		features = append(features, feature)
	}

	err = rows.Err()

	if err != nil {
		return nil, err
	}

	fc := pip.GeoJSONFeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}

	return &fc, nil
}

func (i *SpatialiteIndex) GetIntersectsByPath(path geom.Path, f filter.Filter) ([]spr.StandardPlacesResults, error) {

	db := i.database

	conn, err := db.Conn()

	if err != nil {
		return nil, err
	}

	pending := path.Length()
	points := make([]string, pending)

	for i, c := range path.Vertices() {
		points[i] = fmt.Sprintf("%0.6f %0.6f", c.X, c.Y)
	}

	wkt := fmt.Sprintf("LINESTRING(%s)", strings.Join(points, ","))

	q := fmt.Sprintf("SELECT id FROM geometries WHERE ST_Intersects(GeomFromText('%s'), geom)", wkt)

	rows, err := conn.Query(q)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	places := make([]spr.StandardPlacesResult, 0)

	for rows.Next() {

		var str_id string
		err = rows.Scan(&str_id)

		if err != nil {
			return nil, err
		}

		fc, err := i.cache.Get(str_id)

		if err != nil {
			return nil, err
		}

		s := fc.SPR()

		err = filter.FilterSPR(f, s)

		if err != nil {
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

	results := []spr.StandardPlacesResults{&r}
	return results, nil
}
