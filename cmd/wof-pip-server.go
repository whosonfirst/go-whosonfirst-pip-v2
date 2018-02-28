package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-http-mapzenjs"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-pip/app"
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/flags"
	"github.com/whosonfirst/go-whosonfirst-pip/http"
	"github.com/whosonfirst/go-whosonfirst-pip/index"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"io"
	"io/ioutil"
	// golog "log"
	gohttp "net/http"
	"os"
	"os/signal"
	"runtime"
	godebug "runtime/debug"
	"time"
)

func main() {

	// not at all convinced this (multiple flags sets) is a good way
	// to do this yet but I am still just trying to figure out flags
	// in general so it will do for now... (20180228/thisisaaronland)

	// this is a thing you can do... (20180228/thisisaaronland)
	// fl := flag.Lookup("verbose")
	// golog.Println("FLAG", fl)

	spatialite_flags := flag.NewFlagSet("spatialite", flag.PanicOnError)
	var spatialite_dsn = spatialite_flags.String("dsn", ":memory:", "...")

	polylines_flags := flag.NewFlagSet("polylines", flag.PanicOnError)
	var polylines_coords = polylines_flags.Int("coords", 100, "...")

	www_flags := flag.NewFlagSet("www", flag.PanicOnError)
	var www_path = www_flags.String("path", "/debug", "...")
	// var www_apikey = www_flags.String("api-key", "nextzen-xxxxxx", "...")

	fs_flags := flag.NewFlagSet("fs", flag.PanicOnError)
	var fs_path = fs_flags.String("path", "...", "...")

	var fs_args flags.KeyValueArgs
	flag.Var(&fs_args, "fs", "(0) or more user-defined '{KEY}={VALUE}' arguments to pass to the fs cache")

	var spatialite_args flags.KeyValueArgs
	flag.Var(&spatialite_args, "spatialite", "(0) or more user-defined '{KEY}={VALUE}' arguments to pass to the spatialite database")

	var exclude flags.Exclude
	flag.Var(&exclude, "exclude", "Exclude (WOF) records based on their existential flags. Valid options are: ceased, deprecated, not-current, superseded.")

	var polylines_args flags.KeyValueArgs
	flag.Var(&polylines_args, "polylines", "(0) or more user-defined '{KEY}={VALUE}' arguments to pass to ... polylines")

	var www_args flags.KeyValueArgs
	flag.Var(&www_args, "www", "(0) or more user-defined '{KEY}={VALUE}' arguments to pass to ... www")

	// normal flags

	var host = flag.String("host", "localhost", "The hostname to listen for requests on")
	var port = flag.Int("port", 8080, "The port number to listen for requests on")

	var pip_index = flag.String("index", "rtree", "Valid options are: rtree, spatialite")
	var pip_cache = flag.String("cache", "gocache", "Valid options are: gocache, fs, spatialite")

	var mode = flag.String("mode", "files", "...")
	var procs = flag.Int("processes", runtime.NumCPU()*2, "...")

	var is_wof = flag.Bool("is-wof", true, "...")

	var enable_geojson = flag.Bool("enable-geojson", false, "Allow users to request GeoJSON FeatureCollection formatted responses.")
	var enable_extras = flag.Bool("enable-extras", false, "")
	var enable_candidates = flag.Bool("enable-candidates", false, "")
	var enable_polylines = flag.Bool("enable-polylines", false, "")
	var enable_www = flag.Bool("enable-www", false, "")

	var verbose = flag.Bool("verbose", false, "")

	flag.Parse()

	logger := log.SimpleWOFLogger()
	level := "status"

	if *verbose {
		level = "debug"
	}

	stdout := io.Writer(os.Stdout)
	logger.AddLogger(stdout, level)

	// see above

	spatialite_flags.Parse(spatialite_args.ToFlags())
	polylines_flags.Parse(spatialite_args.ToFlags())
	www_flags.Parse(www_args.ToFlags())
	fs_flags.Parse(fs_args.ToFlags())

	runtime.GOMAXPROCS(*procs)

	if *enable_www {
		logger.Status("-www flag is true causing the following flags to also be true: -enable-geojson -enable-candidates")
		*enable_geojson = true
		*enable_candidates = true
	}

	// cloned from wof-pip.go

	var db *database.SQLiteDatabase

	var appindex index.Index
	var appindex_err error

	var appcache cache.Cache
	var appcache_err error

	logger.Info("index is %s cache is %s", *pip_index, *pip_cache)

	if *pip_index == "spatialite" {

		logger.Debug("setting up spatialite database")
		logger.Debug("spatialite driver is %s and dsn is %s", *pip_index, *spatialite_dsn)

		d, err := database.NewDBWithDriver(*pip_index, *spatialite_dsn)

		if err != nil {
			logger.Fatal("Failed to create spatialite database, because %s", err)
		}

		err = d.LiveHardDieFast()

		if err != nil {
			logger.Fatal("Failed to create spatialite database, because %s", err)
		}

		db = d
	}

	logger.Debug("setting up application cache")

	switch *pip_cache {

	case "gocache":

		opts, err := cache.DefaultGoCacheOptions()

		if err != nil {
			appcache_err = err
		} else {
			appcache, appcache_err = cache.NewGoCache(opts)
		}

	case "fs":
		appcache, appcache_err = cache.NewFSCache(*fs_path)
	case "sqlite":
		appcache, appcache_err = cache.NewSQLiteCache(db)
	case "spatialite":
		appcache, appcache_err = cache.NewSQLiteCache(db)
	default:
		appcache_err = errors.New("Invalid cache layer")
	}

	if appcache_err != nil {
		logger.Fatal("Failed to create caching layer because %s", appcache_err)
	}

	logger.Debug("setting up application index")

	switch *pip_index {
	case "rtree":
		appindex, appindex_err = index.NewRTreeIndex(appcache)
	case "spatialite":
		appindex, appindex_err = index.NewSpatialiteIndex(db, appcache)
	default:
		appindex_err = errors.New("Invalid engine")
	}

	if appindex_err != nil {
		logger.Fatal("failed to create index because %s", appindex_err)
	}

	// end of cloned from...

	indexer_opts, err := app.DefaultApplicationIndexerOptions()

	if err != nil {
		logger.Fatal("failed to create indexer options, because %s", err)
	}

	indexer_opts.IndexMode = *mode
	indexer_opts.IsWOF = *is_wof

	// extras...

	var extras_dsn string

	if *enable_extras {

		index_extras := true

		logger.Debug("setting up extras support")

		// we are relying on the fact that all of these things have already
		// been vetted above and that the spatialite DB in fact has a geojson
		// table (20180228/thisisaaronland)

		// the problem with this approach is that we might be using a SQLite
		// database that was *generated* by the cache/sqlite.go code whose Set()
		// method only knows about cache.CacheItem thingies which don't have a
		// full WOF properties hash so things like '?extras=geom:longitude'
		// will always fail... (20180228/thisisaaronland)

		// for example, this:
		// ./bin/wof-pip-server -index spatialite -cache spatialite -spatialite dsn=test3.db -enable-extras
		//
		// where test3.db has previously been created by doing (something like) this:
		// ./bin/wof-pip -index spatialite -cache spatialite -spatialite dsn=test3.db -mode repo /usr/local/data/whosonfirst-data
		//
		// which will have populated the 'geojson' table in 'test3.db' using the cache.Set()
		// method described above, and which will be lacking a full (WOF) properties
		// dictionary
		//
		// possible solutions include:
		//
		// 1. testing for and using a '-extras dsn=foo.db' flag which has the perverse
		//    side-effect of requiring *two* SQLite databases
		// 2. testing the '-spatialite dsn=foo.db' database for a record that contains
		//    something we know will be in the WOF properties hash but is _not_ part of
		//    the SPR interface (geom:latitude for example) and throwing an error if it
		//    is missing
		// 3. changing the name of the table that the sqlite.Cache Get() method uses and
		//    adding a flag (flags) to query the correct table and... I am having trouble
		//    keeping track of it as I write these words
		//
		// (2) plus proper documentation is probably the easiest thing going forward under
		// the assumption that almost no one is going to be creating *fresh* databases and
		// instead just using the databases that WOF itself produces (20180228/thisisaaronland)

		if *pip_cache == "spatialite" || *pip_cache == "sqlite" {

			dsn := *spatialite_dsn

			// see above - this is solution (2) which is pretty WOF-specific in that it
			// tests for a geom:latitude property which will probably break things if
			// someone is indexing not-WOF documents but we'll just file that as a
			// known-known for now (20180228/thisisaaronland)

			if dsn != ":memory:" {

				db_test, err := database.NewDB(dsn)

				if err != nil {
					logger.Fatal("Failed to set up extras, because %s", err)
				}

				defer db_test.Close()

				conn, err := db_test.Conn()

				if err != nil {
					logger.Fatal("Failed to set up extras, because %s", err)
				}

				row := conn.QueryRow("SELECT body FROM geojson LIMIT 1")

				var body []byte
				err = row.Scan(&body)

				switch {
				case err == sql.ErrNoRows:
					logger.Fatal("Failed to set up extras, because %s", err)
				case err != nil:
					logger.Fatal("Failed to set up extras, because %s", err)
				default:
					// pass
				}

				geom_lat := gjson.GetBytes(body, "properties.geom:latitude")

				if !geom_lat.Exists() {
					logger.Fatal("Failed to set up extras because the first record in the 'geojson' table lacks a 'geom:latitude' property")
				}

				db_test.Close()

				index_extras = false
				extras_dsn = dsn
			}
		}

		if index_extras {

			dsn := *spatialite_dsn

			// MAYBE REVISIT THIS DECISION? (20180228/thisisaaronland)

			if dsn == ":memory:" {

				tmpfile, err := ioutil.TempFile("", "pip-extras")

				if err != nil {
					logger.Fatal("Failed to create temp file, because %s", err)
				}

				tmpfile.Close()
				tmpnam := tmpfile.Name()

				logger.Status("create temporary extras database '%s'", tmpnam)
				dsn = tmpnam

				cleanup := func() {

					logger.Status("remove temporary extras database '%s'", tmpnam)

					err := os.Remove(tmpnam)

					if err != nil {
						logger.Warning("failed to remove %s, because %s", tmpnam, err)
					}
				}

				defer cleanup()

				signal_ch := make(chan os.Signal, 1)
				signal.Notify(signal_ch, os.Interrupt)

				go func() {
					<-signal_ch
					cleanup()
				}()
			}

			extras_dsn = dsn
		}

		logger.Debug("enable extras with dsn %s", extras_dsn)
		logger.Debug("enable extras with indexing %t", index_extras)

		indexer_opts.IndexExtras = index_extras
		indexer_opts.ExtrasDB = extras_dsn
	}

	for _, e := range exclude {

		switch e {
		case "deprecated":
			indexer_opts.IncludeDeprecated = false
		case "ceased":
			indexer_opts.IncludeCeased = false
		case "superseded":
			indexer_opts.IncludeSuperseded = false
		case "not-current":
			indexer_opts.IncludeNotCurrent = false
		default:
			logger.Warning("unknown exclude filter (%s), ignoring", e)
		}
	}

	indexer, err := app.NewApplicationIndexer(appindex, indexer_opts)

	// note: this is "-mode spatialite" not "-engine spatialite"

	if *mode != "spatialite" {

		go func() {

			// TO DO: put this somewhere so that it can be triggered by signal(s)
			// to reindex everything in bulk or incrementally

			t1 := time.Now()

			err = indexer.IndexPaths(flag.Args())

			if err != nil {
				logger.Fatal("failed to index paths because %s", err)
			}

			t2 := time.Since(t1)

			logger.Status("finished indexing in %v", t2)
			godebug.FreeOSMemory()
		}()

		// set up some basic monitoring and feedback stuff

		go func() {

			c := time.Tick(1 * time.Second)

			for _ = range c {

				if !indexer.IsIndexing() {
					continue
				}

				logger.Status("indexing %d records indexed", indexer.Indexed)
			}
		}()
	}

	go func() {

		tick := time.Tick(1 * time.Minute)

		for _ = range tick {
			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)
			logger.Status("memstats system: %8d inuse: %8d released: %8d objects: %6d", ms.HeapSys, ms.HeapInuse, ms.HeapReleased, ms.HeapObjects)
		}
	}()

	// set up the HTTP endpoint

	logger.Debug("setting up intersects handler")

	intersects_opts := http.NewDefaultIntersectsHandlerOptions()
	intersects_opts.EnableGeoJSON = *enable_geojson
	intersects_opts.EnableExtras = *enable_extras
	intersects_opts.ExtrasDB = extras_dsn

	intersects_handler, err := http.IntersectsHandler(appindex, indexer, intersects_opts)

	if err != nil {
		logger.Fatal("failed to create PIP handler because %s", err)
	}

	ping_handler, err := http.PingHandler()

	if err != nil {
		logger.Fatal("failed to create Ping handler because %s", err)
	}

	mux := gohttp.NewServeMux()

	mux.Handle("/ping", ping_handler)
	mux.Handle("/", intersects_handler)

	if *enable_candidates {

		logger.Debug("setting up candidates handler")

		candidateshandler, err := http.CandidatesHandler(appindex, indexer)

		if err != nil {
			logger.Fatal("failed to create Spatial handler because %s", err)
		}

		mux.Handle("/candidates", candidateshandler)
	}

	if *enable_polylines {

		logger.Debug("setting up polylines handler")

		poly_opts := http.NewDefaultPolylineHandlerOptions()
		poly_opts.MaxCoords = *polylines_coords
		poly_opts.EnableGeoJSON = *enable_geojson

		poly_handler, err := http.PolylineHandler(appindex, indexer, poly_opts)

		if err != nil {
			logger.Fatal("failed to create polyline handler because %s", err)
		}

		mux.Handle("/polyline", poly_handler)
	}

	if *enable_www {

		logger.Debug("setting up www handler")

		var www_handler gohttp.Handler

		bundled_handler, err := http.BundledWWWHandler()

		if err != nil {
			logger.Fatal("failed to create (bundled) www handler because %s", err)
		}

		www_handler = bundled_handler

		/*

			mapzenjs_opts := mapzenjs.DefaultMapzenJSOptions()
			mapzenjs_opts.APIKey = *www_apikey

			mapzenjs_handler, err := mapzenjs.MapzenJSHandler(www_handler, mapzenjs_opts)

			if err != nil {
				logger.Fatal("failed to create mapzen.js handler because %s", err)
			}

				mzjs_opts := mapzenjs.DefaultMapzenJSOptions()
				mzjs_opts.APIKey = *api_key

				mzjs_handler, err := mapzenjs.MapzenJSHandler(www_handler, mzjs_opts)

				if err != nil {
					logger.Fatal("failed to create API key handler because %s", err)
				}

				opts := rewrite.DefaultRewriteRuleOptions()

				rewrite_path := *www_path

				if strings.HasSuffix(rewrite_path, "/") {
					rewrite_path = strings.TrimRight(rewrite_path, "/")
				}

				rule := rewrite.RemovePrefixRewriteRule(rewrite_path, opts)
				rules := []rewrite.RewriteRule{rule}

				debug_handler, err := rewrite.RewriteHandler(rules, apikey_handler)

				if err != nil {
					logger.Fatal("failed to create www handler because %s", err)
				}
		*/

		mapzenjs_assets_handler, err := mapzenjs.MapzenJSAssetsHandler()

		if err != nil {
			logger.Fatal("failed to create mapzenjs_assets handler because %s", err)
		}

		mux.Handle("/javascript/mapzen.min.js", mapzenjs_assets_handler)
		mux.Handle("/javascript/tangram.min.js", mapzenjs_assets_handler)
		mux.Handle("/javascript/mapzen.js", mapzenjs_assets_handler)
		mux.Handle("/javascript/tangram.js", mapzenjs_assets_handler)
		mux.Handle("/css/mapzen.js.css", mapzenjs_assets_handler)
		mux.Handle("/tangram/refill-style.zip", mapzenjs_assets_handler)

		mux.Handle("/javascript/mapzen.whosonfirst.pip.js", www_handler)
		mux.Handle("/javascript/slippymap.crosshairs.js", www_handler)
		mux.Handle("/css/mapzen.whosonfirst.pip.css", www_handler)

		mux.Handle(*www_path, www_handler)
	}

	endpoint := fmt.Sprintf("%s:%d", *host, *port)
	logger.Status("listening for requests on %s", endpoint)

	err = gohttp.ListenAndServe(endpoint, mux)

	if err != nil {
		logger.Fatal("failed to start server because %s", err)
	}

	os.Exit(0)
}
