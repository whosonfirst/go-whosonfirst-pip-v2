package index

// https://gist.github.com/simonw/91a1157d1f45ab305c6f48c4ca344de8

import (
	"github.com/skelterjohn/geom"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/geometry"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-pip"
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	"github.com/whosonfirst/go-whosonfirst-spr"
	// golog "log"
	"sync"
)

type SpatialiteIndex struct {
     Index
     Logger *log.WOFLogger
     cache cache.Cache
}

func NewSpatialiteIndex(c cache.Cache) (Index, error) {

	logger := log.SimpleWOFLogger("index")
	
     i := SpatialiteIndex {
       cache: c,
       Logger: logger,
     }

     return &i, nil
}

func (i *SpatialiteIndex) Cache() cache.Cache {

}

func (i *SpatialiteIndex) IndexFeature(f geojson.Feature) error {

}

func GetIntersectsByCoord(geom.Coord, filter.Filter) (spr.StandardPlacesResults, error) {

}

func GetCandidatesByCoord(geom.Coord) (*pip.GeoJSONFeatureCollection, error) {

}

func GetIntersectsByPath(geom.Path, filter.Filter) ([]spr.StandardPlacesResults, error) {

}
