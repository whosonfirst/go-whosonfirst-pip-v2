package main

/*

00:47:14.786746 [wof-pip-server][STATUS] listening on localhost:8080
00:47:56.446116 [wof-pip-server][index][STATUS] SEARCH [-73.51, -73.51]x[45.56, 45.56] 166.853µs
00:47:56.456283 [wof-pip-server][failover][STATUS] GET 1108963037
00:47:56.456394 [wof-pip-server][lru][STATUS] GET 1108963037 9.266µs
00:47:56.468770 [wof-pip-server][failover][STATUS] PRIMARY HIT 1108963037
00:47:56.492490 [wof-pip-server][failover][STATUS] GET 1108963037 12.434419ms
00:47:56.494290 [wof-pip-server][failover][STATUS] GET 1108963041
00:47:56.494369 [wof-pip-server][lru][STATUS] GET 1108963041 7.497µs
00:47:56.494892 [wof-pip-server][failover][STATUS] PRIMARY HIT 1108963041
00:47:56.494941 [wof-pip-server][failover][STATUS] GET 1108963041 602.407µs
00:47:56.495801 [wof-pip-server][failover][STATUS] GET 85875721
00:47:56.495852 [wof-pip-server][lru][STATUS] GET 85875721 4.058µs
00:47:56.496284 [wof-pip-server][failover][STATUS] PRIMARY MISS 85875721
00:47:56.498285 [wof-pip-server][source][STATUS] GET 85875721 1.932644ms
00:47:56.544526 [wof-pip-server][failover][STATUS] SECONDARY HIT 85875721
00:47:56.552458 [wof-pip-server][failover][STATUS] GET 85875721 48.787771ms
00:47:56.564666 [wof-pip-server][index][STATUS] INFLATE {45.557093 -73.513641} 108.347259ms
00:47:56.565009 [wof-pip-server][index][STATUS] INTERSECT {45.557093 -73.513641} 119.109189ms

curl -s 'localhost:8080?latitude=45.557093&longitude=-73.513641' | python -mjson.tool
{
    "places": [
        {
            "mz:is_ceased": 0,
            "mz:is_current": -1,
            "mz:is_deprecated": 0,
            "mz:is_superseded": 0,
            "mz:is_superseding": 0,
            "mz:uri": "https://whosonfirst.mapzen.com/data/110/896/303/7/1108963037.geojson",
            "wof:country": "CA",
            "wof:id": 1108963037,
            "wof:name": "Saint Lawrence River",
            "wof:parent_id": -3,
            "wof:path": "110/896/303/7/1108963037.geojson",
            "wof:placetype": "neighbourhood",
            "wof:repo": "whosonfirst-data",
            "wof:superseded_by": [],
            "wof:supersedes": []
        },
        {
            "mz:is_ceased": 0,
            "mz:is_current": 0,
            "mz:is_deprecated": 1,
            "mz:is_superseded": 1,
            "mz:is_superseding": 0,
            "mz:uri": "https://whosonfirst.mapzen.com/data/858/757/21/85875721.geojson",
            "wof:country": "CA",
            "wof:id": 85875721,
            "wof:name": "Vieux Longueuil",
            "wof:parent_id": 101738793,
            "wof:path": "858/757/21/85875721.geojson",
            "wof:placetype": "neighbourhood",
            "wof:repo": "whosonfirst-data",
            "wof:superseded_by": [
                1108961051
            ],
            "wof:supersedes": []
        }
    ]
}

*/

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
	// "sync"
	// "syscall"
	"time"
)

func main() {

	var host = flag.String("host", "localhost", "The hostname to listen for requests on")
	var port = flag.Int("port", 8080, "The port number to listen for requests on")

	var cache = flag.String("cache", "lru", "...")

	var failover_cache = flag.String("failover-cache", "lru", "...")

	var lru_cache_size = flag.Int("lru-cache-size", 1024, "")
	var lru_cache_trigger = flag.Int("lru-cache-trigger", 0, "")

	var mode = flag.String("mode", "files", "")
	var procs = flag.Int("processes", runtime.NumCPU()*2, "")

	var not_wof = flag.Bool("not-wof", false, "")

	var debug = flag.Bool("debug", false, "")
	var as_geojson = flag.Bool("debug-as-geojson", false, "")

	var api_key = flag.String("api-key", "mapzen-xxxxxxx", "")

	flag.Parse()

	runtime.GOMAXPROCS(*procs)

	logger := log.SimpleWOFLogger()

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
	default:
		logger.Fatal("Invalid cache layer %s", *cache)
	}

	appcache_opts.LRUCacheSize = *lru_cache_size
	appcache_opts.LRUCacheTriggerSize = *lru_cache_trigger
	appcache_opts.SourceCacheRoot = "/usr/local/data"

	appcache, err := app.ApplicationCache(appcache_opts)

	if err != nil {
		logger.Fatal("Failed to creation application cache, because %s", err)
	}

	logger.Status("Use %s cache layer w/ cache size %d", *cache, appcache.Size())

	appindex, err := app.ApplicationIndex(appcache)

	if err != nil {
		logger.Fatal("failed to create index because %s", err)
	}

	indexer_opts, err := app.DefaultApplicationIndexerOptions()

	if err != nil {
		logger.Fatal("failed to create indexer options because %s", err)
	}

	// TO DO: put this somewhere so that it can be triggered by signal(s)
	// to reindex everything in bulk or incrementally

	indexer_opts.IndexMode = *mode

	if *not_wof {
		indexer_opts.IsWOF = false
	}

	indexer, err := app.NewApplicationIndexer(appindex, indexer_opts)

	go func() {

		err = indexer.IndexPaths(flag.Args())

		if err != nil {
			logger.Fatal("failed to index paths because %s", err)
		}

		logger.Status("finished indexing")
		godebug.FreeOSMemory()
	}()

	intersects_opts := http.NewDefaultIntersectsHandlerOptions()
	intersects_opts.AsGeoJSON = *as_geojson

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

	if *debug {

		mapzenjs_handler, err := mapzenjs.MapzenJSHandler()

		if err != nil {
			logger.Fatal("failed to create mapzen.js handler because %s", err)
		}

		www_handler, err := http.WWWHandler()

		if err != nil {
			logger.Fatal("failed to create www handler because %s", err)
		}

		fs := http.WWWFileSystem()

		apikey_handler, err := mapzenjs.MapzenAPIKeyHandler(www_handler, fs, *api_key)

		if err != nil {
			logger.Fatal("failed to create query handler because %s", err)
		}

		opts := rewrite.DefaultRewriteRuleOptions()

		rule := rewrite.RemovePrefixRewriteRule("/debug", opts)
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
		mux.Handle("/css/mapzen.js.css", mapzenjs_handler)
		mux.Handle("/tangram/refill-style.zip", mapzenjs_handler)

		mux.Handle("/javascript/mapzen.whosonfirst.pip.js", www_handler)
		mux.Handle("/css/mapzen.whosonfirst.pip.css", www_handler)

		mux.Handle("/debug/", debug_handler)
	}

	mux.Handle("/ping", ping_handler)
	mux.Handle("/", intersects_handler)

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

	go func() {
		tick := time.Tick(1 * time.Minute)

		for _ = range tick {

			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)
			logger.Status("System: %8d Inuse: %8d Released: %8d Objects: %6d\n", ms.HeapSys, ms.HeapInuse, ms.HeapReleased, ms.HeapObjects)
		}
	}()

	err = gracehttp.Serve(&gohttp.Server{Addr: endpoint, Handler: mux})

	if err != nil {
		logger.Fatal("failed to start server because %s", err)
	}

	os.Exit(0)
}
