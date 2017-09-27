package index

import (
	"github.com/skelterjohn/geom"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-pip"
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	"github.com/whosonfirst/go-whosonfirst-spr"
)

type Index interface {
	IndexFeature(geojson.Feature) error
	Cache() cache.Cache
	GetIntersectsByCoord(geom.Coord, filter.Filter) (spr.StandardPlacesResults, error)
	GetCandidatesByCoord(geom.Coord) (*pip.GeoJSONFeatureCollection, error)
	GetIntersectsForCoords([]geom.Coord, filter.Filter) (spr.StandardPlacesResults, error)
}

type Candidate interface{} // mmmmmaybe?
