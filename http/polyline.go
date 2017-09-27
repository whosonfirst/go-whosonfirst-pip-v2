package http

// https://developers.google.com/maps/documentation/utilities/polylineutility

import (
	"encoding/json"
	"github.com/skelterjohn/geom"
	"github.com/twpayne/go-polyline"
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

		if str_polyline == "" {
			gohttp.Error(rsp, "Missing 'polyline' parameter", gohttp.StatusBadRequest)
			return
		}

		polyline_coords, _, err := polyline.DecodeCoords([]byte(str_polyline))

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusBadRequest)
			return
		}

		if len(polyline_coords) > opts.MaxCoords {
			gohttp.Error(rsp, "E_EXCESSIVE_COORDINATES", gohttp.StatusBadRequest)
			return
		}
		
		coords := make([]geom.Coord, 0)

		for _, latlon := range polyline_coords {

			lat := latlon[0]
			lon := latlon[1]
			
			c, err := geojson_utils.NewCoordinateFromLatLons(lat, lon)

			if err != nil {
				gohttp.Error(rsp, err.Error(), gohttp.StatusBadRequest)
				return
			}

			coords = append(coords, c)
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
