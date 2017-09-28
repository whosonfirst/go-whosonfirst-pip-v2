package pip

// would that we could have a single all-purpose geojson interface
// but since we don't necessarily know what *kind* of geometry we're
// dealing with (I mean here we do but that's not the point) it's
// not clear what we say gets returned by a Coordinates() method in
// a Geometry interface - is it really even that important? I suppose
// it would be nice to be able to have GetCandidatesByCoord return
// an interface-thingy but for now this will do (20170822/thisisaaronland)

// to whit: what is the relationship between this and all the GeoJSON
// structs in cache/utils.go... I am not sure yet (20170921/thisisaaronland)

type GeoJSONPoint []float64

type GeoJSONRing []GeoJSONPoint

type GeoJSONPolygon []GeoJSONRing

type GeoJSONMultiPolygon []GeoJSONPolygon

type GeoJSONGeometry struct {
	Type        string              `json:"type"`
	Coordinates GeoJSONMultiPolygon `json:"coordinates"`
}

type GeoJSONProperties interface{}

type GeoJSONFeature struct {
	Type       string            `json:"type"`
	Geometry   GeoJSONGeometry   `json:"geometry"`
	Properties GeoJSONProperties `json:"properties"`
}

type GeoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

type GeoJSONFeatureCollectionSet struct {
	Type        string                      `json:"type"`
	Collections []*GeoJSONFeatureCollection `json:"features"`
}
