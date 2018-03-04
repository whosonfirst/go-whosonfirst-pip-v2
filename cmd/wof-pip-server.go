package main

import (
	"fmt"
	"github.com/whosonfirst/go-http-mapzenjs"
	"github.com/whosonfirst/go-http-rewrite"
	"github.com/whosonfirst/go-whosonfirst-pip/app"
	"github.com/whosonfirst/go-whosonfirst-pip/flags"
	"github.com/whosonfirst/go-whosonfirst-pip/http"
	"log"
	gohttp "net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

func main() {

	fl, err := flags.CommonFlags()

	if err != nil {
		log.Fatal(err)
	}

	fl.String("host", "localhost", "The hostname to listen for requests on")
	fl.Int("port", 8080, "The port number to listen for requests on")

	fl.Bool("enable-geojson", false, "Allow users to request GeoJSON FeatureCollection formatted responses.")
	fl.Bool("enable-candidates", false, "")
	fl.Bool("enable-polylines", false, "")
	fl.Bool("enable-www", false, "")

	fl.Int("polylines-coords", 100, "...")
	fl.String("www-path", "/debug", "...")
	fl.String("www-api-key", "xxxxxx", "...")

	flags.Parse(fl)

	pip, err := app.NewPIPApplication(fl)

	if err != nil {
		pip.Logger.Fatal("Failed to create new PIP application, because %s", err)
	}

	pip_index, _ := flags.StringVar(fl, "index")
	pip_cache, _ := flags.StringVar(fl, "cache")
	mode, _ := flags.StringVar(fl, "mode")

	pip.Logger.Info("index is %s cache is %s mode is %s", pip_index, pip_cache, mode)

	err = pip.IndexPaths(fl.Args())

	if err != nil {
		pip.Logger.Fatal("Failed to index paths, because %s", err)
	}

	go func() {

		tick := time.Tick(1 * time.Minute)

		for _ = range tick {
			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)
			pip.Logger.Status("memstats system: %8d inuse: %8d released: %8d objects: %6d", ms.HeapSys, ms.HeapInuse, ms.HeapReleased, ms.HeapObjects)
		}
	}()

	// set up the HTTP endpoint

	enable_www, _ := flags.BoolVar(fl, "enable-www")

	if enable_www {
		pip.Logger.Status("-enable-www flag is true causing the following flags to also be true: -enable-geojson -enable-candidates")

		fl.Set("enable-geojson", "true")
		fl.Set("enable-candidates", "true")
	}

	pip.Logger.Debug("setting up intersects handler")

	enable_geojson, _ := flags.BoolVar(fl, "enable-geojson")

	intersects_opts := http.NewDefaultIntersectsHandlerOptions()
	intersects_opts.EnableGeoJSON = enable_geojson

	intersects_handler, err := http.IntersectsHandler(pip.Index, pip.Indexer, pip.Extras, intersects_opts)

	if err != nil {
		pip.Logger.Fatal("failed to create PIP handler because %s", err)
	}

	ping_handler, err := http.PingHandler()

	if err != nil {
		pip.Logger.Fatal("failed to create Ping handler because %s", err)
	}

	mux := gohttp.NewServeMux()

	mux.Handle("/ping", ping_handler)
	mux.Handle("/", intersects_handler)

	enable_candidates, _ := flags.BoolVar(fl, "enable-candidates")
	enable_polylines, _ := flags.BoolVar(fl, "enable-polylines")

	if enable_candidates {

		pip.Logger.Debug("setting up candidates handler")

		candidateshandler, err := http.CandidatesHandler(pip.Index, pip.Indexer)

		if err != nil {
			pip.Logger.Fatal("failed to create Spatial handler because %s", err)
		}

		mux.Handle("/candidates", candidateshandler)
	}

	if enable_polylines {

		pip.Logger.Debug("setting up polylines handler")

		poly_coords, _ := flags.IntVar(fl, "polylines-coords")

		poly_opts := http.NewDefaultPolylineHandlerOptions()
		poly_opts.MaxCoords = poly_coords
		poly_opts.EnableGeoJSON = enable_geojson

		poly_handler, err := http.PolylineHandler(pip.Index, pip.Indexer, poly_opts)

		if err != nil {
			pip.Logger.Fatal("failed to create polyline handler because %s", err)
		}

		mux.Handle("/polyline", poly_handler)
	}

	if enable_www {

		www_path, _ := flags.StringVar(fl, "www-path")
		api_key, _ := flags.StringVar(fl, "www-api-key")

		pip.Logger.Debug("setting up www handler at %s", www_path)

		www_handler, err := http.BundledWWWHandler()

		if err != nil {
			pip.Logger.Fatal("failed to create (bundled) www handler because %s", err)
		}

		// all the HTML-y bits expect everything to be hanging off of '/' but the default
		// www endpoint is '/debug' so we set up an internal rewrite handler here
		// (20180304/thisisaaronland)

		mapzenjs_assets_handler, err := mapzenjs.MapzenJSAssetsHandler()

		if err != nil {
			pip.Logger.Fatal("failed to create mapzenjs_assets handler because %s", err)
		}

		mapzenjs_opts := mapzenjs.DefaultMapzenJSOptions()
		mapzenjs_opts.APIKey = api_key

		mapzenjs_handler, err := mapzenjs.MapzenJSHandler(www_handler, mapzenjs_opts)

		if err != nil {
			pip.Logger.Fatal("failed to create (bundled) mapzenjs handler because %s", err)
		}

		rewrite_opts := rewrite.DefaultRewriteRuleOptions()
		rewrite_path := www_path

		if strings.HasSuffix(rewrite_path, "/") {
			rewrite_path = strings.TrimRight(rewrite_path, "/")
		}

		rule := rewrite.RemovePrefixRewriteRule(rewrite_path, rewrite_opts)
		rules := []rewrite.RewriteRule{rule}

		rewrite_handler, err := rewrite.RewriteHandler(rules, mapzenjs_handler)

		if err != nil {
			pip.Logger.Fatal("failed to create rewrite handler because %s", err)
		}

		mux.Handle("/javascript/mapzen.min.js", mapzenjs_assets_handler)
		mux.Handle("/javascript/tangram.min.js", mapzenjs_assets_handler)
		mux.Handle("/javascript/mapzen.js", mapzenjs_assets_handler)
		mux.Handle("/javascript/tangram.js", mapzenjs_assets_handler)
		mux.Handle("/css/mapzen.js.css", mapzenjs_assets_handler)
		mux.Handle("/tangram/refill-style.zip", mapzenjs_assets_handler)
		mux.Handle("/tangram/refill-style-themes-label.zip", mapzenjs_assets_handler)

		mux.Handle("/javascript/mapzen.whosonfirst.pip.js", www_handler)
		mux.Handle("/javascript/slippymap.crosshairs.js", www_handler)
		mux.Handle("/css/mapzen.whosonfirst.pip.css", www_handler)

		mux.Handle(www_path, rewrite_handler)
	}

	host, _ := flags.StringVar(fl, "host")
	port, _ := flags.IntVar(fl, "port")

	endpoint := fmt.Sprintf("%s:%d", host, port)
	pip.Logger.Status("listening for requests on %s", endpoint)

	err = gohttp.ListenAndServe(endpoint, mux)

	if err != nil {
		pip.Logger.Fatal("failed to start server because %s", err)
	}

	os.Exit(0)
}
