package http

import (
	"encoding/json"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	wof "github.com/whosonfirst/go-whosonfirst-index"
	pip "github.com/whosonfirst/go-whosonfirst-pip-v2/index"
	_ "log"
	gohttp "net/http"
	"strconv"
)

func CandidatesHandler(i pip.Index, idx *wof.Indexer) (gohttp.Handler, error) {

	fn := func(rsp gohttp.ResponseWriter, req *gohttp.Request) {

		if idx.IsIndexing() {
			gohttp.Error(rsp, "indexing records", gohttp.StatusServiceUnavailable)
			return
		}

		query := req.URL.Query()

		str_lat := query.Get("latitude")
		str_lon := query.Get("longitude")

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

		coord, err := utils.NewCoordinateFromLatLons(lat, lon)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusBadRequest)
			return
		}

		candidates, err := i.GetCandidatesByCoord(coord)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			return
		}

		enc, err := json.Marshal(candidates)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			return
		}

		rsp.Header().Set("Content-Type", "application/json")
		rsp.Header().Set("Access-Control-Allow-Origin", "*")

		rsp.Write(enc)
	}

	h := gohttp.HandlerFunc(fn)
	return h, nil
}
