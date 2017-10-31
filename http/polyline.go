package http

import (
	"encoding/json"
	"github.com/skelterjohn/geom"
	geojson_utils "github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-pip"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	pip_index "github.com/whosonfirst/go-whosonfirst-pip/index"
	pip_utils "github.com/whosonfirst/go-whosonfirst-pip/utils"
	"github.com/whosonfirst/go-whosonfirst-spr"
	"log"
	"math"
	gohttp "net/http"
	"strconv"
)

type PolylineResultsUnique struct {
	spr.StandardPlacesResults `json:",omitempty"`
	Rows                      []spr.StandardPlacesResult `json:"places"`
	Pagination                pip.Pagination             `json:"pagination,omitempty"`
}

func (r *PolylineResultsUnique) Results() []spr.StandardPlacesResult {
	return r.Rows
}

type PolylineHandlerOptions struct {
	AllowGeoJSON bool
	MaxCoords    int
}

func NewDefaultPolylineHandlerOptions() *PolylineHandlerOptions {

	opts := PolylineHandlerOptions{
		AllowGeoJSON: false,
		MaxCoords:    500,
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
		str_unique := query.Get("unique")
		str_format := query.Get("format")

		str_page := query.Get("page")
		str_per_page := query.Get("per_page")

		page := 1
		per_page := opts.MaxCoords

		total_count := 0
		page_count := 1

		if str_polyline == "" {
			gohttp.Error(rsp, "Missing 'polyline' parameter", gohttp.StatusBadRequest)
			return
		}

		if str_format == "geojson" && !opts.AllowGeoJSON {
			gohttp.Error(rsp, "Invalid format", gohttp.StatusBadRequest)
			return
		}

		if str_page != "" {

			p, err := strconv.Atoi(str_page)

			if err != nil {
				gohttp.Error(rsp, err.Error(), gohttp.StatusBadRequest)
			}

			page = p
		}

		if str_per_page != "" {

			p, err := strconv.Atoi(str_per_page)

			if err != nil {
				gohttp.Error(rsp, err.Error(), gohttp.StatusBadRequest)
			}

			per_page = p

			if per_page > opts.MaxCoords {
				gohttp.Error(rsp, "Invalid per_page parameter", gohttp.StatusBadRequest)
			}
		}

		unique := false
		poly_factor := 1.0e5

		if str_valhalla != "" {
			poly_factor = 1.0e6
		}

		if str_unique != "" {
			unique = true
		}

		path, err := DecodePolyline(str_polyline, poly_factor)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusBadRequest)
			return
		}

		log.Println("PATH", path, path.Length())
		
		total_count = path.Length()

		if total_count > per_page {

			first := (page - 1) * per_page
			last := first + per_page

			log.Println("FIRST", first)
			log.Println("LAST", last)
			
			total_count_fl := float64(total_count)
			per_page_fl := float64(per_page)

			page_count_fl := math.Ceil(total_count_fl / per_page_fl)
			page_count = int(page_count_fl)

			vertices := path.Vertices()

			slice := geom.Path{}

			for _, c := range vertices[first:last] {
				slice.AddVertex(c)
			}

			path = &slice
			
			log.Println("PATH", path, path.Length())			
		}

		pagination := pip.Pagination{
			TotalCount: total_count,
			Page:       page,
			PerPage:    per_page,
			PageCount:  page_count,
		}

		log.Println("PAGINATION", pagination)
		
		filters, err := filter.NewSPRFilterFromQuery(query)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusBadRequest)
			return
		}

		results, err := i.GetIntersectsByPath(*path, filters)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			return
		}

		var final interface{}
		final = results

		if unique {

			rows := make([]spr.StandardPlacesResult, 0)
			seen := make(map[string]bool)

			for _, rs := range results {

				for _, r := range rs.Results() {

					id := r.Id()

					_, ok := seen[id]

					if ok {
						continue
					}

					rows = append(rows, r)
					seen[id] = true
				}
			}

			unq := PolylineResultsUnique{
				Rows:       rows,
				Pagination: pagination,
			}

			final = &unq
		}

		if str_format == "geojson" {

			if unique {

				collection, err := pip_utils.ResultsToFeatureCollection(final.(spr.StandardPlacesResults), i)

				if err != nil {
					gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
					return
				}

				collection.Pagination = pagination
				final = collection

			} else {

				collections := make([]*pip.GeoJSONFeatureCollection, 0)

				for _, rs := range results {

					collection, err := pip_utils.ResultsToFeatureCollection(rs, i)

					if err != nil {
						gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
						return
					}

					collections = append(collections, collection)
				}

				collection_set := pip.GeoJSONFeatureCollectionSet{
					Type:        "FeatureCollectionSet",
					Collections: collections,

				}

				final = &collection_set
			}
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

func DecodePolyline(encoded string, f float64) (*geom.Path, error) {

	var count, index int

	tempLatLng := [2]int{0, 0}

	path := geom.Path{}

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

			path.AddVertex(coord)
		}

		count++
	}

	return &path, nil
}
