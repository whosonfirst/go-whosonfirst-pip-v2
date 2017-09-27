package http

import (
	"encoding/json"
	"github.com/skelterjohn/geom"
	geojson_utils "github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"	
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	pip_index "github.com/whosonfirst/go-whosonfirst-pip/index"
	pip_utils "github.com/whosonfirst/go-whosonfirst-pip/utils"
	_ "log"
	gohttp "net/http"
)

type PolylineHandlerOptions struct {
	AsGeoJSON bool
	MaxCoords int
}

func NewDefaultPolylineHandlerOptions() *PolylineHandlerOptions {

	opts := PolylineHandlerOptions{
		AsGeoJSON: false,
		MaxCoords: 500,			   
	}

	return &opts
}

func PolylineHandler(i pip_index.Index, idx *index.Indexer, opts *PolylineHandlerOptions) (gohttp.Handler, error) {

	fn := func(rsp gohttp.ResponseWriter, req *gohttp.Request) {

		if idx.IsIndexing() {
			gohttp.Error(rsp, "indexing records", gohttp.StatusServiceUnavailable)
			return
		}

		query := req.URL.Query()
		str_polyline := query.Get("polyline")
		str_valhalla := query.Get("valhalla")		

		if str_polyline == "" {
			gohttp.Error(rsp, "Missing 'polyline' parameter", gohttp.StatusBadRequest)
			return
		}

		poly_factor := 1.0e5
		
		if str_valhalla != "" {
		   	poly_factor = 1.0e6
		}		

		coords, err := DecodePolyline(str_polyline, poly_factor)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusBadRequest)
			return
		}

		if len(coords) > opts.MaxCoords {
			gohttp.Error(rsp, "E_EXCESSIVE_COORDINATES", gohttp.StatusBadRequest)
			return
		}
		
		filters, err := filter.NewSPRFilterFromQuery(query)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusBadRequest)
			return
		}

		results, err := i.GetIntersectsForCoords(coords, filters)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			return
		}

		var final interface{}
		final = results

		// note: we will need a suitable function to handle polyline responses
		// once said response has been formalized (20170927/thisisaaronland)

		if opts.AsGeoJSON {

			collection, err := pip_utils.ResultsToFeatureCollection(results, i)

			if err != nil {
				gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
				return
			}

			final = collection
		}

		js, err := json.Marshal(final)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			return
		}

		rsp.Header().Set("Content-Type", "application/json")
		rsp.Header().Set("Access-Control-Allow-Origin", "*")

		rsp.Write(js)
	}

	h := gohttp.HandlerFunc(fn)
	return h, nil
}

// the DecodePolyline function is cribbed from Paul Mach's NewPathFromEncoding function here:
// https://github.com/paulmach/go.geo/blob/master/path.go
//
// We don't need to import the rest of the package just the code that can handle decoding
// plain-vanilla GOOG 5-decimal point polylines as well as Valhalla's 6-decimal point lines
// defined here: https://mapzen.com/documentation/mobility/decoding/
//
// see also: https://developers.google.com/maps/documentation/utilities/polylineutility
// (20170927/thisisaaronland)

func DecodePolyline(encoded string, f float64) ([]geom.Coord, error) {
	
	var count, index int

     	coords := make([]geom.Coord, 0)
	tempLatLng := [2]int{0, 0}

	for index < len(encoded) {
		var result int
		var b = 0x20
		var shift uint

		for b >= 0x20 {
			b = int(encoded[index]) - 63
			index++

			result |= (b & 0x1f) << shift
			shift += 5
		}

		// sign dection
		if result&1 != 0 {
			result = ^(result >> 1)
		} else {
			result = result >> 1
		}

		if count%2 == 0 {
			result += tempLatLng[0]
			tempLatLng[0] = result
		} else {
			result += tempLatLng[1]
			tempLatLng[1] = result

			lon := float64(tempLatLng[1]) / f
			lat := float64(tempLatLng[0]) / f

			coord, err := geojson_utils.NewCoordinateFromLatLons(lat, lon)

			if err != nil {
				return nil, err
			}

			coords = append(coords, coord)
		}

		count++
	}

	return coords, nil
}
