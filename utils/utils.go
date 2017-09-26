package utils

import (
	"context"
	geojson_utils "github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-pip"
	pip_index "github.com/whosonfirst/go-whosonfirst-pip/index"
	"github.com/whosonfirst/go-whosonfirst-spr"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io"
	"io/ioutil"
	_ "log"
	"strconv"
)

// https://github.com/whosonfirst/go-whosonfirst-geojson/blob/master/geojson.go#L27-L38
// I don't know... it was 2015 (20170922/thisisaaronland)

type V1WOFSpatial struct {
	Id         int64
	Name       string
	Placetype  string
	Offset     int
	Deprecated bool
	Superseded bool
}

func IsWOFRecord(fh io.Reader) (bool, error) {

	body, err := ioutil.ReadAll(fh)

	if err != nil {
		return false, err
	}

	possible := []string{
		"properties.wof:id",
	}

	id := geojson_utils.Int64Property(body, possible, -1)

	if id == -1 {
		return false, nil
	}

	return true, nil
}

func IsValidRecord(fh io.Reader, ctx context.Context) (bool, error) {

	path, err := index.PathForContext(ctx)

	if err != nil {
		return false, err
	}

	if path == index.STDIN {
		return true, nil
	}

	is_wof, err := uri.IsWOFFile(path)

	if err != nil {
		return false, err
	}

	if !is_wof {
		return false, nil
	}

	is_alt, err := uri.IsAltFile(path)

	if err != nil {
		return false, err
	}

	if is_alt {
		return false, nil
	}

	return true, nil
}

// basically we need this in order to roll over all the servers/services
// without any downtime (20170922/thisisaaronland)

func ResultsToV1Results(results spr.StandardPlacesResults) ([]*V1WOFSpatial, error) {

	spatial := make([]*V1WOFSpatial, 0)

	for _, r := range results.Results() {

		str_id := r.Id()

		id, err := strconv.ParseInt(str_id, 10, 64)

		if err != nil {
			id = -1
		}

		d := r.IsDeprecated()
		s := r.IsSuperseded()

		deprecated := false
		superseded := false

		if d.IsTrue() && d.IsKnown() {
			deprecated = true
		}

		if s.IsTrue() && s.IsKnown() {
			superseded = true
		}

		sp := V1WOFSpatial{
			Id:         id,
			Name:       r.Name(),
			Placetype:  r.Placetype(),
			Offset:     -1,
			Deprecated: deprecated,
			Superseded: superseded,
		}

		spatial = append(spatial, &sp)
	}

	return spatial, nil
}

func ResultsToFeatureCollection(results spr.StandardPlacesResults, idx pip_index.Index) (*pip.GeoJSONFeatureCollection, error) {

	cache := idx.Cache()

	features := make([]pip.GeoJSONFeature, 0)

	for _, r := range results.Results() {

		fc, err := cache.Get(r.Id())

		if err != nil {
			return nil, err
		}

		f := pip.GeoJSONFeature{
			Type:       "Feature",
			Properties: fc.SPR(),
			Geometry:   fc.Geometry(),
		}

		features = append(features, f)
	}

	collection := pip.GeoJSONFeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}

	return &collection, nil
}
