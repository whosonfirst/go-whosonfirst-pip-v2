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

func GetIntersectsByCoord(geom.Coord, filter.Filter) (spr.StandardPlacesResults, error) {

	/*

		select
		  id, placetype, name, length(geom), properties
		from
		  whosonfirst
		where
		  within(GeomFromText('POINT(' || :longitude || ' ' || :latitude || ')'), geom)
		  and rowid in (
		        SELECT pkid FROM idx_whosonfirst_geom
		        where xmin < :longitude
		        and xmax > :longitude
		        and ymin < :latitude
		        and ymax > :latitude)
		order by placetype desc;

	*/

	return nil, errors.New("PLEASE WRITE ME")
}

func GetCandidatesByCoord(geom.Coord) (*pip.GeoJSONFeatureCollection, error) {
	return nil, errors.New("PLEASE WRITE ME")
}

func GetIntersectsByPath(geom.Path, filter.Filter) ([]spr.StandardPlacesResults, error) {
	return nil, errors.New("PLEASE WRITE ME")
}
