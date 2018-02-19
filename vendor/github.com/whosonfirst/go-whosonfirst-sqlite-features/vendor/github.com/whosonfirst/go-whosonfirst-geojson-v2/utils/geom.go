package utils

import (
	"github.com/skelterjohn/geom"
)

func NewCoordinateFromLatLons(lat float64, lon float64) (geom.Coord, error) {

	coord := new(geom.Coord)

	coord.Y = lat
	coord.X = lon

	return *coord, nil
}

func NewRectFromLatLons(minlat float64, minlon float64, maxlat float64, maxlon float64) (geom.Rect, error) {

	bbox := new(geom.Rect)

	min_coord, err := NewCoordinateFromLatLons(minlat, minlon)

	if err != nil {
		return *bbox, err
	}

	max_coord, err := NewCoordinateFromLatLons(maxlat, maxlon)

	if err != nil {
		return *bbox, err
	}

	bbox.Min = min_coord
	bbox.Max = max_coord

	return *bbox, nil
}

func NewPolygonFromCoords(coords []geom.Coord) (geom.Polygon, error) {

	path := geom.Path{}

	for _, c := range coords {
		path.AddVertex(c)
	}

	poly := new(geom.Polygon)
	poly.Path = path

	return *poly, nil
}
