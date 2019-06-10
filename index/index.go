package index

import (
	"github.com/skelterjohn/geom"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-pip-v2"
	"github.com/whosonfirst/go-whosonfirst-pip-v2/cache"
	"github.com/whosonfirst/go-whosonfirst-pip-v2/filter"
	"github.com/whosonfirst/go-whosonfirst-spr"
)

type Index interface {
	IndexFeature(geojson.Feature) error
	Cache() cache.Cache
	Close() error
	GetIntersectsByCoord(geom.Coord, filter.Filter) (spr.StandardPlacesResults, error)
	GetCandidatesByCoord(geom.Coord) (*pip.GeoJSONFeatureCollection, error)
	GetIntersectsByPath(geom.Path, filter.Filter) ([]spr.StandardPlacesResults, error)
}

type Candidate interface{} // mmmmmaybe?
