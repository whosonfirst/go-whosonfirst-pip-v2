package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	pip_index "github.com/whosonfirst/go-whosonfirst-pip/index"
	pip_utils "github.com/whosonfirst/go-whosonfirst-pip/utils"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"log"
	gohttp "net/http"
	"strconv"
	"strings"
	"sync"
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

					db, err := database.NewDB(opts.ExtrasDB)

					if err != nil {
						gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
						return

					}

					defer db.Close()

					conn, err := db.Conn()

					if err != nil {
						gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
						return
					}

					js, err, _ = AppendExtras(js, extras, conn)

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

func AppendExtras(js []byte, extras []string, conn *sql.DB) ([]byte, error, bool) {

	type update struct {
		Index int
		Body  []byte
	}

	done_ch := make(chan bool)
	update_ch := make(chan update)
	error_ch := make(chan error)

	rsp := gjson.GetBytes(js, "places.#")
	places := rsp.Array()

	count := len(places)
	log.Println("PLACES", count)
	
	for i, pl := range places {

		go func(idx int, pl gjson.Result) {

			defer func() {
				done_ch <- true
			}()

			spr := []byte(pl.Raw)

			log.Println("UPDATE RECORD %d", i)

			updated, err := AppendExtrasToSPRBytes(spr, extras, conn)

			if err != nil {
				error_ch <- err
				return
			}

			up := update{
				Index: i,
				Body:  updated,
			}

			update_ch <- up

		}(i, pl)
	}

	mu := new(sync.Mutex)
	remaining := count

	for remaining > 0 {

		select {
		case <-done_ch:
			log.Println("DONE", remaining)
			remaining -= 1
		case err := <-error_ch:
			log.Println("ERROR", err, remaining)
			remaining -= 1
		case up := <-update_ch:

			mu.Lock()
			set_path := fmt.Sprintf("places.%d", up.Index)
			log.Println("UPDATE", set_path, up.Body)

			js, _ = sjson.SetBytes(js, set_path, up.Body)
			mu.Unlock()
		}
	}

	log.Println(string(js))

	return js, nil, true
}

func AppendExtrasToSPRBytes(spr []byte, extras []string, conn *sql.DB) ([]byte, error) {

	rsp := gjson.GetBytes(spr, "wof:id")

	if !rsp.Exists() {
		return nil, errors.New("Unable to determine wof:id")
	}

	wofid := rsp.Int()

	log.Println("APPEND", wofid)

	// apparently JSON_EXTRACT isn't available in go-sqlite yet?
	// 2017/12/17 20:07:00 420561633 no such function: JSON_EXTRACT
	// row := conn.QueryRow("SELECT JSON_EXTRACT(feature, '$.properties') FROM geojson WHERE id=?", wofid)

	// see also: https://github.com/whosonfirst/go-whosonfirst-pip-v2/issues/19

	row := conn.QueryRow("SELECT body FROM geojson WHERE id=?", wofid)

	var body []byte
	err := row.Scan(&body)

	switch {
	case err == sql.ErrNoRows:
		return spr, err
	case err != nil:
		return spr, err
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

		log.Println("PATHS", paths)

		for _, p := range paths {

			// see above inre absence of JSON_EXTRACT function

			get_path := fmt.Sprintf("properties.%s", p)
			set_path := fmt.Sprintf("%s", p)

			log.Println("GET", wofid, get_path)
			log.Println("SET", wofid, set_path)

			v := gjson.GetBytes(body, get_path)

			if v.Exists() {
				spr, err = sjson.SetBytes(spr, set_path, v.Value())
			} else {
				spr, err = sjson.SetBytes(spr, set_path, nil)
			}

			if err != nil {
				return spr, err
			}
		}
	}

	return spr, nil
}
