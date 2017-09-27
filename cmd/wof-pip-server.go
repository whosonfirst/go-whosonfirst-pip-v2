package main

import (
	"flag"
	"fmt"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/whosonfirst/go-http-mapzenjs"
	"github.com/whosonfirst/go-http-rewrite"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-pip/app"
	"github.com/whosonfirst/go-whosonfirst-pip/http"
	gohttp "net/http"
	"os"
	"runtime"
	godebug "runtime/debug"
	"strings"
	// "sync"
	// "syscall"
	"time"
)

func main() {

	var host = flag.String("host", "localhost", "The hostname to listen for requests on")
	var port = flag.Int("port", 8080, "The port number to listen for requests on")

	var cache = flag.String("cache", "gocache", "...")
	var cache_all = flag.Bool("cache-all", false, "")

	var failover_cache = flag.String("failover-cache", "lru", "...")

	var lru_cache_size = flag.Int("lru-cache-size", 1024, "...")
	var lru_cache_trigger = flag.Int("lru-cache-trigger", 2000, "")

	var source_cache_root = flag.String("source-cache-root", "/usr/local/data", "...")

	var mode = flag.String("mode", "files", "...")
	var procs = flag.Int("processes", runtime.NumCPU()*2, "...")

	var plain_old_geojson = flag.Bool("plain-old-geojson", false, "...")

	var www = flag.Bool("www", false, "")
	var www_path = flag.String("www-path", "/debug/", "")
	var www_geojson = flag.Bool("www-as-geojson", false, "")
	var www_local = flag.Bool("www-local", false, "")
	var www_local_root = flag.String("www-local-root", "", "")

	var api_key = flag.String("mapzen-api-key", "mapzen-xxxxxxx", "")

	var polylines = flag.Bool("polylines", false, "")
	var polylines_coords = flag.Int("polylines-max-coords", 500, "")

	flag.Parse()

	runtime.GOMAXPROCS(*procs)

	logger := log.SimpleWOFLogger()

	// set up the caching layer for the point-in-poly index

	appcache_opts, err := app.DefaultApplicationCacheOptions()

	if err != nil {
		logger.Fatal("Failed to creation application cache options, because %s", err)
	}

	appcache_opts.IndexMode = *mode
	appcache_opts.IndexPaths = flag.Args()

	switch *cache {
	case "lru":
		appcache_opts.LRUCache = true
	case "failover":
		appcache_opts.FailoverCache = true
		appcache_opts.FailoverCacheEngine = *failover_cache
	case "gocache":
		appcache_opts.GoCache = true
	case "source":
		appcache_opts.SourceCache = true
	default:
		logger.Fatal("Invalid cache layer %s", *cache)
	}

	appcache_opts.LRUCacheSize = *lru_cache_size
	appcache_opts.LRUCacheTriggerSize = *lru_cache_trigger
	appcache_opts.SourceCacheRoot = *source_cache_root

	if *cache_all {
		appcache_opts.LRUCacheSize = 0
		appcache_opts.LRUCacheTriggerSize = 0
	}

	if *plain_old_geojson {
		appcache_opts.IsWOF = false
	}

	appcache, err := app.ApplicationCache(appcache_opts)

	if err != nil {
		logger.Fatal("Failed to creation application cache, because %s", err)
	}

	// set up the actual point-in-poly index

	appindex, err := app.ApplicationIndex(appcache)

	if err != nil {
		logger.Fatal("failed to create index because %s", err)
	}

	// set up the index (all these records) thingy

	indexer_opts, err := app.DefaultApplicationIndexerOptions()

	if err != nil {
		logger.Fatal("failed to create indexer options because %s", err)
	}

	indexer_opts.IndexMode = *mode

	if *plain_old_geojson {
		indexer_opts.IsWOF = false // if true we skip the WOF specific "is valid record" checks
	}

	indexer, err := app.NewApplicationIndexer(appindex, indexer_opts)

	go func() {

		// TO DO: put this somewhere so that it can be triggered by signal(s)
		// to reindex everything in bulk or incrementally

		err = indexer.IndexPaths(flag.Args())

		if err != nil {
			logger.Fatal("failed to index paths because %s", err)
		}

		logger.Status("finished indexing")
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

	go func() {

		tick := time.Tick(1 * time.Minute)

		for _ = range tick {
			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)
			logger.Status("memstats system: %8d inuse: %8d released: %8d objects: %6d", ms.HeapSys, ms.HeapInuse, ms.HeapReleased, ms.HeapObjects)
		}
	}()

	// set up the HTTP endpoint

	intersects_opts := http.NewDefaultIntersectsHandlerOptions()
	intersects_opts.AsGeoJSON = *www_geojson

	intersects_handler, err := http.IntersectsHandler(appindex, indexer, intersects_opts)

	if err != nil {
		logger.Fatal("failed to create PIP handler because %s", err)
	}

	ping_handler, err := http.PingHandler()

	if err != nil {
		logger.Fatal("failed to create Ping handler because %s", err)
	}

	endpoint := fmt.Sprintf("%s:%d", *host, *port)
	logger.Status("listening on %s", endpoint)

	mux := gohttp.NewServeMux()

	mux.Handle("/ping", ping_handler)
	mux.Handle("/", intersects_handler)

	if *polylines {

		poly_opts := http.NewDefaultPolylineHandlerOptions()
		poly_opts.MaxCoords = *polylines_coords

		poly_handler, err := http.PolylineHandler(appindex, indexer, poly_opts)

		if err != nil {
			logger.Fatal("failed to create polyline handler because %s", err)
		}

		mux.Handle("/polyline", poly_handler)
	}

	if *www {

		mapzenjs_handler, err := mapzenjs.MapzenJSHandler()

		if err != nil {
			logger.Fatal("failed to create mapzen.js handler because %s", err)
		}

		var www_handler gohttp.Handler
		var www_fs gohttp.FileSystem

		if *www_local {

			local_fs, err := http.LocalWWWFileSystem(*www_local_root)

			if err != nil {
				logger.Fatal("failed to create (local) file system because %s", err)
			}

			local_handler, err := http.LocalWWWHandler(local_fs)

			if err != nil {
				logger.Fatal("failed to create (local) www handler because %s", err)
			}

			www_handler = local_handler
			www_fs = local_fs

		} else {

			bundled_handler, err := http.BundledWWWHandler()

			if err != nil {
				logger.Fatal("failed to create (bundled) www handler because %s", err)
			}

			bundled_fs, err := http.BundledWWWFileSystem()

			if err != nil {
				logger.Fatal("failed to create (bundled) file system because %s", err)
			}

			www_handler = bundled_handler
			www_fs = bundled_fs
		}

		apikey_handler, err := mapzenjs.MapzenAPIKeyHandler(www_handler, www_fs, *api_key)

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

		candidateshandler, err := http.CandidatesHandler(appindex, indexer)

		if err != nil {
			logger.Fatal("failed to create Spatial handler because %s", err)
		}

		mux.Handle("/candidates", candidateshandler)

		mux.Handle("/javascript/mapzen.min.js", mapzenjs_handler)
		mux.Handle("/javascript/tangram.min.js", mapzenjs_handler)
		mux.Handle("/javascript/mapzen.js", mapzenjs_handler)
		mux.Handle("/javascript/tangram.js", mapzenjs_handler)
		mux.Handle("/css/mapzen.js.css", mapzenjs_handler)
		mux.Handle("/tangram/refill-style.zip", mapzenjs_handler)

		mux.Handle("/javascript/mapzen.whosonfirst.pip.js", www_handler)
		mux.Handle("/css/mapzen.whosonfirst.pip.css", www_handler)

		mux.Handle(*www_path, debug_handler)
	}

	// make it go

	err = gracehttp.Serve(&gohttp.Server{Addr: endpoint, Handler: mux})

	if err != nil {
		logger.Fatal("failed to start server because %s", err)
	}

	os.Exit(0)
}
