package http

import (
	"encoding/json"
	"github.com/skelterjohn/geom"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-pip"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	pip_index "github.com/whosonfirst/go-whosonfirst-pip/index"
	pip_utils "github.com/whosonfirst/go-whosonfirst-pip/utils"
	"github.com/whosonfirst/go-whosonfirst-spr"
	_ "log"
	"math"
	gohttp "net/http"
	"strconv"
)

// see this - see the way some things are an SPR thingy and both things have a `pip`
// Pagination property but not an `spr` equivalent - yeah, we need to sort that out
// (20171031/thisisaaronland)

type PolylineResults struct {
	// spr.StandardPlacesResults `json:",omitempty"`
	Rows       [][]spr.StandardPlacesResult `json:"places"`
	Pagination pip.Pagination               `json:"pagination,omitempty"`
}

type PolylineResultsUnique struct {
	// spr.StandardPlacesResults `json:",omitempty"`
	Rows       [][]spr.StandardPlacesResult `json:"places"`
	Pagination pip.Pagination               `json:"pagination,omitempty"`
}

// see above inre `pip` and `spr` and things left to figure out...
// (20171031/thisisaaronland)
// func (r *PolylineResultsUnique) Results() []spr.StandardPlacesResult {
//	return r.Rows
// }

type PolylineHandlerOptions struct {
	EnableGeoJSON bool
	MaxCoords     int
}

func NewDefaultPolylineHandlerOptions() *PolylineHandlerOptions {

	opts := PolylineHandlerOptions{
		EnableGeoJSON: false,
		MaxCoords:     100,
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
		str_precision := query.Get("precision")
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

		if str_format == "geojson" && !opts.EnableGeoJSON {
			gohttp.Error(rsp, "Invalid format", gohttp.StatusBadRequest)
			return
		}

		if str_page != "" {

			p, err := strconv.Atoi(str_page)

			if err != nil {
				gohttp.Error(rsp, err.Error(), gohttp.StatusBadRequest)
				return
			}

			if p < 1 {
				gohttp.Error(rsp, "Invalid page parameter", gohttp.StatusBadRequest)
				return
			}

			page = p
		}

		if str_per_page != "" {

			p, err := strconv.Atoi(str_per_page)

			if err != nil {
				gohttp.Error(rsp, err.Error(), gohttp.StatusBadRequest)
				return
			}

			per_page = p

			if per_page < 1 {
				gohttp.Error(rsp, "Invalid per_page value", gohttp.StatusBadRequest)
				return
			}

			if per_page > opts.MaxCoords {
				gohttp.Error(rsp, "Invalid per_page value", gohttp.StatusBadRequest)
				return
			}
		}

		unique := false
		poly_factor := 1.0e5

		if str_precision != "" {

			f, err := pip_utils.StringPrecisionToFactor(str_precision)

			if err != nil {
				gohttp.Error(rsp, "Invalid precision value", gohttp.StatusBadRequest)
				return
			}

			poly_factor = f
		}

		if str_unique != "" {
			unique = true
		}

		path, err := pip_utils.DecodePolyline(str_polyline, poly_factor)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusBadRequest)
			return
		}

		total_count = path.Length()

		if total_count > per_page {

			// log.Println("PAGE", page, per_page)

			first := (page - 1) * per_page
			last := first + per_page

			// log.Println("SLICE", first, last)

			if last > total_count {
				last = total_count - 1
			}

			total_count_fl := float64(total_count)
			per_page_fl := float64(per_page)

			page_count_fl := math.Ceil(total_count_fl / per_page_fl)
			page_count = int(page_count_fl)

			if page > page_count {
				gohttp.Error(rsp, "Invalid page parameter", gohttp.StatusBadRequest)
				return
			}

			vertices := path.Vertices()

			slice := geom.Path{}

			for _, c := range vertices[first:last] {
				slice.AddVertex(c)
			}

			path = &slice
		}

		pagination := pip.Pagination{
			TotalCount: total_count,
			Page:       page,
			PerPage:    per_page,
			PageCount:  page_count,
		}

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

		p_rows := make([][]spr.StandardPlacesResult, 0)

		for _, rs := range results {

			rows := make([]spr.StandardPlacesResult, 0)

			for _, r := range rs.Results() {
				rows = append(rows, r)
			}

			p_rows = append(p_rows, rows)
		}

		p_results := PolylineResults{
			Rows:       p_rows,
			Pagination: pagination,
		}

		var final interface{}
		final = p_results

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

			p_rows := [][]spr.StandardPlacesResult{rows}

			unq := PolylineResultsUnique{
				Rows:       p_rows,
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
