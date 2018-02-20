package index

// https://gist.github.com/simonw/91a1157d1f45ab305c6f48c4ca344de8

import (
	"errors"
	"github.com/skelterjohn/geom"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	// "github.com/whosonfirst/go-whosonfirst-geojson-v2/geometry"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-pip"
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	"github.com/whosonfirst/go-whosonfirst-spr"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"github.com/whosonfirst/go-whosonfirst-sqlite/utils"
	// golog "log"
	// "sync"
)

type SpatialiteIndex struct {
	Index
	Logger   *log.WOFLogger
	database *database.SQLiteDatabase
	cache    cache.Cache
}

func NewSpatialiteIndex(db *database.SQLiteDatabase, c cache.Cache) (Index, error) {

	logger := log.SimpleWOFLogger("index")

	ok_geom, err := utils.HasTable(db, "geometries")

	if err != nil {
		return nil, err
	}

	if !ok_geom {
		return nil, errors.New("Missing 'geometries' table")
	}

	i := SpatialiteIndex{
		database: db,
		cache:    c,
		Logger:   logger,
	}

	return &i, nil
}

func (i *SpatialiteIndex) Cache() cache.Cache {
	return i.cache
}

func (i *SpatialiteIndex) IndexFeature(f geojson.Feature) error {

	return nil
}

func (i *SpatialiteIndex) GetIntersectsByCoord(coord geom.Coord, f filter.Filter) (spr.StandardPlacesResults, error) {

	db := i.database

	conn, err := db.Conn()

	if err != nil {
		return nil, err
	}

	lat := coord.Y
	lon := coord.X

	// spatialite> select name from states where within(GeomFromText('POINT(-97.74342 30.26771)'),states.Geometry);

	q := `SELECT id FROM geometries WHERE WITHIN(GeomFromText('POINT(? ?)'), geom) AND rowid IN (
		SELECT pkid FROM idx_whosonfirst_geom
		        where xmin < ?
		        and xmax > ?
		        and ymin < ?
		        and ymax > ?
        )`

	rows, err := conn.Query(q, lon, lat, lon, lon, lat, lat)

	if err != nil {
		return nil, err
	}

	i.Logger.Status("%v", rows)

	return nil, errors.New("PLEASE WRITE ME")
}

func (i *SpatialiteIndex) GetCandidatesByCoord(coord geom.Coord) (*pip.GeoJSONFeatureCollection, error) {
	return nil, errors.New("PLEASE WRITE ME")
}

func (i *SpatialiteIndex) GetIntersectsByPath(path geom.Path, f filter.Filter) ([]spr.StandardPlacesResults, error) {
	return nil, errors.New("PLEASE WRITE ME")
}
