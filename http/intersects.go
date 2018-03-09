package http

import (
	"encoding/json"
	geojson_utils "github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	wof_index "github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-pip/extras"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	"github.com/whosonfirst/go-whosonfirst-pip/index"
	"github.com/whosonfirst/go-whosonfirst-pip/utils"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	gohttp "net/http"
	"strconv"
	"strings"
)

type IntersectsHandlerOptions struct {
	EnableGeoJSON bool
}

func NewDefaultIntersectsHandlerOptions() *IntersectsHandlerOptions {

	opts := IntersectsHandlerOptions{
		EnableGeoJSON: false,
	}

	return &opts
}

func IntersectsHandler(i index.Index, idx *wof_index.Indexer, extras_db *database.SQLiteDatabase, opts *IntersectsHandlerOptions) (gohttp.Handler, error) {

	fn := func(rsp gohttp.ResponseWriter, req *gohttp.Request) {

		if idx.IsIndexing() {
			gohttp.Error(rsp, "indexing records", gohttp.StatusServiceUnavailable)
			return
		}

		query := req.URL.Query()

		str_lat := query.Get("latitude")
		str_lon := query.Get("longitude")
		str_format := query.Get("format")

		v1 := query.Get("v1")

		if str_format == "geojson" && !opts.EnableGeoJSON {
			gohttp.Error(rsp, "Invalid format", gohttp.StatusBadRequest)
			return
		}

		if str_lat == "" {
			gohttp.Error(rsp, "Missing 'latitude' parameter", gohttp.StatusBadRequest)
			return
		}

		if str_lon == "" {
			gohttp.Error(rsp, "Missing 'longitude' parameter", gohttp.StatusBadRequest)
			return
		}

		lat, err := strconv.ParseFloat(str_lat, 64)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusBadRequest)
			return
		}

		lon, err := strconv.ParseFloat(str_lon, 64)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusBadRequest)
			return
		}

		coord, err := geojson_utils.NewCoordinateFromLatLons(lat, lon)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusBadRequest)
			return
		}

		filters, err := filter.NewSPRFilterFromQuery(query)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusBadRequest)
			return
		}

		results, err := i.GetIntersectsByCoord(coord, filters)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			return
		}

		var final interface{}
		final = results

		if v1 != "" {

			v1_results, err := utils.ResultsToV1Results(results)

			if err != nil {
				gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
				return
			}

			final = v1_results

		} else if str_format == "geojson" {

			collection, err := utils.ResultsToFeatureCollection(results, i)

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

		// experimental - see notes in extras/extras.go (20180303/thisisaaronland)

		if extras_db != nil {

			var extras_paths []string

			str_extras := query.Get("extras")
			str_extras = strings.Trim(str_extras, " ")

			if str_extras != "" {
				extras_paths = strings.Split(str_extras, ",")
			}

			if len(extras_paths) > 0 {

				js, err = extras.AppendExtrasWithSPRResults(js, results, extras_paths, extras_db)

				if err != nil {
					gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
					return
				}
			}
		}

		rsp.Header().Set("Content-Type", "application/json")
		rsp.Header().Set("Access-Control-Allow-Origin", "*")

		rsp.Write(js)
	}

	h := gohttp.HandlerFunc(fn)
	return h, nil
}
