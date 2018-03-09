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

	fs, err := flags.CommonFlags()

	if err != nil {
		log.Fatal(err)
	}

	err = flags.AppendWWWFlags(fs)

	flags.Parse(fs)

	err = flags.ValidateCommonFlags(fs)

	if err != nil {
		log.Fatal(err)
	}

	err = flags.ValidateWWWFlags(fs)

	if err != nil {
		log.Fatal(err)
	}

	pip, err := app.NewPIPApplication(fs)

	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to create new PIP application, because %s", err))
	}

	pip_index, _ := flags.StringVar(fs, "index")
	pip_cache, _ := flags.StringVar(fs, "cache")
	mode, _ := flags.StringVar(fs, "mode")

	pip.Logger.Info("index is %s cache is %s mode is %s", pip_index, pip_cache, mode)

	err = pip.IndexPaths(fs.Args())

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

	pip.Logger.Debug("setting up intersects handler")

	enable_geojson, _ := flags.BoolVar(fs, "enable-geojson")

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

	enable_www, _ := flags.BoolVar(fs, "enable-www")
	enable_candidates, _ := flags.BoolVar(fs, "enable-candidates")
	enable_polylines, _ := flags.BoolVar(fs, "enable-polylines")

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

		poly_coords, _ := flags.IntVar(fs, "polylines-max-coords")

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

		www_path, _ := flags.StringVar(fs, "www-path")
		api_key, _ := flags.StringVar(fs, "www-api-key")

		pip.Logger.Debug("setting up www handler at %s", www_path)

		www_handler, err := http.BundledWWWHandler()

		if err != nil {
			pip.Logger.Fatal("failed to create (bundled) www handler because %s", err)
		}

		// squirt an API key in to document.body in HTML pages

		mapzenjs_opts := mapzenjs.DefaultMapzenJSOptions()
		mapzenjs_opts.APIKey = api_key

		mapzenjs_handler, err := mapzenjs.MapzenJSHandler(www_handler, mapzenjs_opts)

		if err != nil {
			pip.Logger.Fatal("failed to create (bundled) mapzenjs handler because %s", err)
		}

		// all the HTML-y bits expect everything to be hanging off of '/' but the default
		// www endpoint is '/debug' so we set up an internal rewrite handler here
		// (20180304/thisisaaronland)

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

		mapzenjs_assets_handler, err := mapzenjs.MapzenJSAssetsHandler()

		if err != nil {
			pip.Logger.Fatal("failed to create mapzenjs_assets handler because %s", err)
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

	host, _ := flags.StringVar(fs, "host")
	port, _ := flags.IntVar(fs, "port")

	endpoint := fmt.Sprintf("%s:%d", host, port)
	pip.Logger.Status("listening for requests on %s", endpoint)

	err = gohttp.ListenAndServe(endpoint, mux)

	if err != nil {
		pip.Logger.Fatal("failed to start server because %s", err)
	}

	os.Exit(0)
}
