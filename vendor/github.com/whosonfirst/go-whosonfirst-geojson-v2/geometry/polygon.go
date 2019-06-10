package geometry

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/skelterjohn/geom"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	_ "log"
)

type Polygon struct {
	geojson.Polygon `json:",omitempty"`
	Exterior        geom.Polygon   `json:"exterior"`
	Interior        []geom.Polygon `json:"interior"`
}

func (p Polygon) ExteriorRing() geom.Polygon {
	return p.Exterior
}

func (p Polygon) InteriorRings() []geom.Polygon {
	return p.Interior
}

func (p Polygon) ContainsCoord(c geom.Coord) bool {

	ext := p.ExteriorRing()

	contains := false

	if ext.ContainsCoord(c) {

		contains = true

		for _, int := range p.InteriorRings() {

			if int.ContainsCoord(c) {
				contains = false
				break
			}
		}
	}

	return contains
}

func GeometryForFeature(f geojson.Feature) (*geojson.Geometry, error) {

	// see notes below in PolygonsForFeature

	t := gjson.GetBytes(f.Bytes(), "geometry.type")

	if !t.Exists() {
		return nil, errors.New("Failed to determine geometry.type")
	}

	c := gjson.GetBytes(f.Bytes(), "geometry.coordinates")

	if !c.Exists() {
		return nil, errors.New("Failed to determine geometry.coordinates")
	}

	g := gjson.GetBytes(f.Bytes(), "geometry")

	var geom geojson.Geometry
	err := gjson.Unmarshal([]byte(g.Raw), &geom)

	if err != nil {
		return nil, err
	}

	return &geom, nil
}

func PolygonsForFeature(f geojson.Feature) ([]geojson.Polygon, error) {

	// so here's the thing - in the first function (GeometryForFeature)
	// we're going UnMarshal the geomtry and then in the second function
	// (PolygonsForGeometry) we Marshal it again - we do this because the
	// second function hands off to a bunch of 'gjsonToFoo' functions to
	// create a bunch of of geojson.Polygon thingies... which can't be
	// round-tripped to and from serialized blobs because the underlying
	// geom package (specifically the polygon/path package) doesn't export
	// public coordinates... which becomes a problem when you are trying
	// to cache geojson thingies in go-whosonfirst-pip using a caching thing
	// that writes to disk (or a database) and so have no way to rebuild
	// the geometries (see above wrt public coordinates)... which means
	// that we're probably going to change the interface for cache.Feature
	// in go-whosonfirst-pip to require a blob of []byte rather than a set
	// of []geojson.Polygon... so that's why we're two-stepping the existing
	// bytes in f, mostly so that there is a generic GeometryForFeature
	// function that caching layers can access... I mean I suppose we could
	// monkey-patch the geom package too but not today
	// (20170920/thisisaaronland)

	g, err := GeometryForFeature(f)

	if err != nil {
		return nil, err
	}

	return PolygonsForGeometry(g)
}

func PolygonsForGeometry(g *geojson.Geometry) ([]geojson.Polygon, error) {

	b, err := json.Marshal(g)

	if err != nil {
		return nil, err
	}

	t := gjson.GetBytes(b, "type")
	c := gjson.GetBytes(b, "coordinates")

	coords := c.Array()

	if len(coords) == 0 {
		return nil, errors.New("Invalid geometry.coordinates")
	}

	polys := make([]geojson.Polygon, 0)

	switch t.String() {

	case "LineString":

		multi_coords := make([]geom.Coord, len(coords))

		for i, c := range coords {

			pt := c.Array()

			lat := pt[1].Float()
			lon := pt[0].Float()

			coord, _ := utils.NewCoordinateFromLatLons(lat, lon)
			multi_coords[i] = coord
		}

		exterior, err := utils.NewPolygonFromCoords(multi_coords)

		if err != nil {
			return nil, err
		}

		polygon := Polygon{
			Exterior: exterior,
		}

		polys = []geojson.Polygon{polygon}

	case "Polygon":

		// c === rings (below)

		polygon, err := gjson_coordsToPolygon(c)

		if err != nil {
			return nil, err
		}

		polys = append(polys, polygon)

	case "MultiPolygon":

		for _, rings := range coords {

			polygon, err := gjson_coordsToPolygon(rings)

			if err != nil {
				return nil, err
			}

			polys = append(polys, polygon)
		}

	case "Point":

		lat := coords[1].Float()
		lon := coords[0].Float()

		coord, _ := utils.NewCoordinateFromLatLons(lat, lon)

		coords := []geom.Coord{
			coord, coord,
			coord, coord,
			coord,
		}

		exterior, err := utils.NewPolygonFromCoords(coords)

		if err != nil {
			return nil, err
		}

		interior := make([]geom.Polygon, 0)

		polygon := Polygon{
			Exterior: exterior,
			Interior: interior,
		}

		polys = []geojson.Polygon{polygon}
		return polys, nil

	case "MultiPoint":

		multi_coords := make([]geom.Coord, len(coords))

		for i, c := range coords {

			pt := c.Array()

			lat := pt[1].Float()
			lon := pt[0].Float()

			coord, _ := utils.NewCoordinateFromLatLons(lat, lon)
			multi_coords[i] = coord
		}

		exterior, err := utils.NewPolygonFromCoords(multi_coords)

		if err != nil {
			return nil, err
		}

		polygon := Polygon{
			Exterior: exterior,
		}

		polys = []geojson.Polygon{polygon}

	default:

		msg := fmt.Sprintf("Invalid geometry type '%s'", t.String())
		return nil, errors.New(msg)
	}

	return polys, nil
}

func gjson_coordsToPolygon(r gjson.Result) (geojson.Polygon, error) {

	rings := r.Array()

	count_rings := len(rings)
	count_interior := count_rings - 1

	exterior, err := gjson_linearRingToGeomPolygon(rings[0])

	if err != nil {
		return nil, err
	}

	interior := make([]geom.Polygon, count_interior)

	for i := 1; i <= count_interior; i++ {

		poly, err := gjson_linearRingToGeomPolygon(rings[i])

		if err != nil {
			return nil, err
		}

		interior = append(interior, poly)
	}

	polygon := Polygon{
		Exterior: exterior,
		Interior: interior,
	}

	return &polygon, nil
}

func gjson_linearRingToGeomPolygon(r gjson.Result) (geom.Polygon, error) {

	coords := make([]geom.Coord, 0)

	for _, pt := range r.Array() {

		lonlat := pt.Array()

		lat := lonlat[1].Float()
		lon := lonlat[0].Float()

		coord, _ := utils.NewCoordinateFromLatLons(lat, lon)
		coords = append(coords, coord)
	}

	return utils.NewPolygonFromCoords(coords)
}
