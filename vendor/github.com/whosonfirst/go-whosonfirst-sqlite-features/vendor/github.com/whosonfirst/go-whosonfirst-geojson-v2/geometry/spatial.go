package geometry

import (
	"github.com/skelterjohn/geom"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
)

func FeatureContainsCoord(f geojson.Feature, c geom.Coord) (bool, error) {

	polys, err := PolygonsForFeature(f)

	if err != nil {
		return false, err
	}

	return PolygonsContainsCoord(polys, c)
}

func PolygonsContainsCoord(polys []geojson.Polygon, c geom.Coord) (bool, error) {

	contains := false

	for _, p := range polys {

		if p.ContainsCoord(c) {
			contains = true
			break
		}
	}

	return contains, nil
}
