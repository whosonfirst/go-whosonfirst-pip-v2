package http

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	pip_index "github.com/whosonfirst/go-whosonfirst-pip/index"
	pip_utils "github.com/whosonfirst/go-whosonfirst-pip/utils"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	_ "log"
	gohttp "net/http"
	"strconv"
	"strings"
)

type IntersectsHandlerOptions struct {
	AllowGeoJSON bool
	AllowExtras  bool   // see notes below
	ExtrasDB     string // see notes below
}

func NewDefaultIntersectsHandlerOptions() *IntersectsHandlerOptions {

	opts := IntersectsHandlerOptions{
		AllowGeoJSON: false,
		AllowExtras:  false,
		ExtrasDB:     "",
	}

	return &opts
}

func IntersectsHandler(i pip_index.Index, idx *index.Indexer, opts *IntersectsHandlerOptions) (gohttp.Handler, error) {

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

		if str_format == "geojson" && !opts.AllowGeoJSON {
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

		coord, err := utils.NewCoordinateFromLatLons(lat, lon)

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

			v1_results, err := pip_utils.ResultsToV1Results(results)

			if err != nil {
				gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
				return
			}

			final = v1_results

		} else if str_format == "geojson" {

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

		// this is VERY VERY experimental still - basically we're going to decorate the
		// final JSON output with extras data read from one or more SQLite databases -
		// in time this will probably be updated to use go-whosonfirst-readwrite.Reader
		// instances and some "S3 SELECT" -like for user-defined databases but not today
		//
		// (20171217/thisisaaronland)

		if opts.AllowExtras {

			str_extras := query.Get("extras")
			str_extras = strings.Trim(str_extras, " ")

			var extras []string

			if str_extras != "" {
				extras = strings.Split(str_extras, ",")
			}

			if len(extras) > 0 {

				// currently (and maybe ever really) this is only supported for SPR
				// responses - it probably wouldn't be that hard to make it work for
				// geojson feature collection results (20171217/thisisaaronland)

				places := gjson.GetBytes(js, "places.#.wof:id")

				if places.Exists() {

					js, err, _ = AppendExtras(js, extras, places, opts.ExtrasDB)

					if err != nil {
						gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
						return
					}
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

func AppendExtras(js []byte, extras []string, places gjson.Result, db_path string) ([]byte, error, bool) {

	db, err := database.NewDB(db_path)

	if err != nil {
		return js, err, false
	}

	defer db.Close()

	conn, err := db.Conn()

	if err != nil {
		return js, err, false
	}

	for i, id := range places.Array() {

		wofid := id.Int()

		// apparently JSON_EXTRACT isn't available in go-sqlite yet?
		// 2017/12/17 20:07:00 420561633 no such function: JSON_EXTRACT
		// row := conn.QueryRow("SELECT JSON_EXTRACT(body, '$.properties') FROM geojson WHERE id=?", wofid)

		// see also: https://github.com/whosonfirst/go-whosonfirst-pip-v2/issues/19
		
		row := conn.QueryRow("SELECT body FROM geojson WHERE id=?", wofid)

		var body []byte
		err = row.Scan(&body)

		switch {
		case err == sql.ErrNoRows:
			return js, nil, false
		case err != nil:
			return js, err, false
		default:
			// pass
		}

		for _, e := range extras {

			paths := make([]string, 0)

			if strings.HasSuffix(e, "*") || strings.HasSuffix(e, ":") {

				e = strings.Replace(e, "*", "", -1)

				props := gjson.GetBytes(body, "properties")

				for k, _ := range props.Map() {

					if strings.HasPrefix(k, e) {
						paths = append(paths, k)
					}
				}

			} else {
				paths = append(paths, e)
			}

			for _, p := range paths {

				// see above inre absence of JSON_EXTRACT function

				get_path := fmt.Sprintf("properties.%s", p)
				set_path := fmt.Sprintf("places.%d.%s", i, p)

				v := gjson.GetBytes(body, get_path)

				if v.Exists() {
					js, err = sjson.SetBytes(js, set_path, v.Value())
				} else {
					js, err = sjson.SetBytes(js, set_path, nil)
				}

				if err != nil {
					return js, err, false
				}
			}
		}

		break
	}

	return js, nil, true
}
