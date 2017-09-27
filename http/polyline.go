package http

import (
	"encoding/json"
	"github.com/skelterjohn/geom"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	pip_index "github.com/whosonfirst/go-whosonfirst-pip/index"
	pip_utils "github.com/whosonfirst/go-whosonfirst-pip/utils"
	_ "log"
	gohttp "net/http"
)

type PolylineHandlerOptions struct {
	AsGeoJSON bool
}

func NewDefaultPolylineHandlerOptions() *PolylineHandlerOptions {

	opts := PolylineHandlerOptions{
		AsGeoJSON: false,
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

		polyline := query.Get("polyline")

		if polyline == "" {
			gohttp.Error(rsp, "Missing 'polyline' parameter", gohttp.StatusBadRequest)
			return
		}

		gohttp.Error(rsp, "Y U TRY TO USE THIS YET", gohttp.StatusServiceUnavailable)

		// placeholder until we have a polyline-to-coords thing (20170927/thisisaaronland)

		coords := make([]geom.Coord, 0)

		inputs, err := filter.NewSPRInputs()

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusBadRequest)
			return
		}

		inputs.Placetypes = query["placetype"]
		inputs.IsCurrent = query["is_current"]
		inputs.IsDeprecated = query["is_deprecated"]
		inputs.IsCeased = query["is_ceased"]
		inputs.IsSuperseded = query["is_superseded"]
		inputs.IsSuperseding = query["is_superseding"]

		filters, err := filter.NewSPRFilterFromInputs(inputs)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusBadRequest)
			return
		}

		results, err := i.GetIntersectsByPolyline(coords, filters)

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
