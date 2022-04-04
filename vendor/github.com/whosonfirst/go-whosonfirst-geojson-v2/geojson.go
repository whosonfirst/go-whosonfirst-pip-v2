package geojson

import (
	"github.com/skelterjohn/geom"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
)

type Feature interface {
	Id() string
	Name() string
	Placetype() string
	String() string
	Bytes() []byte
	BoundingBoxes() (BoundingBoxes, error)
	Polygons() ([]Polygon, error)
	SPR() (spr.StandardPlacesResult, error)
	ContainsCoord(geom.Coord) (bool, error)
}

type BoundingBoxes interface {
	Bounds() []*geom.Rect
	MBR() geom.Rect
}

type Centroid interface {
	Coord() geom.Coord
	Source() string
	ToString() (string, error)
}

type Polygon interface {
	ExteriorRing() geom.Polygon
	InteriorRings() []geom.Polygon
	ContainsCoord(geom.Coord) bool
}

type Geometry interface{}
