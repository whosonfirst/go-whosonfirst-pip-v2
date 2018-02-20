package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/skelterjohn/geom"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-pip/app"
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	pip "github.com/whosonfirst/go-whosonfirst-pip/index"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	// golog "log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func PIPLatLon(i pip.Index, lat float64, lon float64, f filter.Filter, logger *log.WOFLogger) error {

	c, err := utils.NewCoordinateFromLatLons(lat, lon)

	if err != nil {
		return err
	}

	return PIP(i, c, f, logger)
}

func PIP(i pip.Index, c geom.Coord, f filter.Filter, logger *log.WOFLogger) error {

	t1 := time.Now()

	r, err := i.GetIntersectsByCoord(c, f)

	t2 := time.Since(t1)

	if err != nil {
		return err
	}

	logger.Status("time to count %d records: %v\n", len(r.Results()), t2)

	body, err := json.Marshal(r)

	if err != nil {
		return err
	}

	fmt.Println(string(body))
	return nil
}

func main() {

	var interactive = flag.Bool("interactive", false, "")

	var engine = flag.String("engine", "rtree", "")
	var dsn = flag.String("dsn", ":memory:", "")

	var lat = flag.Float64("latitude", 0.0, "")
	var lon = flag.Float64("longitude", 0.0, "")
	var point = flag.String("point", "", "")

	var mode = flag.String("mode", "files", "")
	var procs = flag.Int("processes", runtime.NumCPU()*2, "")

	/*
		var source_cache_enable = flag.Bool("source-cache", false, "")
		var source_cache_root = flag.String("source-cache-data-root", "/usr/local/data", "")

		var lru_cache_enable = flag.Bool("lru-cache", false, "")
		var lru_cache_size = flag.Int("lru-cache-size", 0, "")
		var lru_cache_trigger = flag.Int("lru-cache-trigger", 0, "")

		var failover_cache_enable = flag.Bool("failover-cache", false, "")
	*/

	flag.Parse()

	runtime.GOMAXPROCS(*procs)

	logger := log.SimpleWOFLogger()

	if *point != "" {

		parts := strings.Split(*point, ",")

		if len(parts) != 2 {
			logger.Fatal("Can not parse point")
		}

		str_lat := strings.Trim(parts[0], " ")
		str_lon := strings.Trim(parts[1], " ")

		fl_lat, err := strconv.ParseFloat(str_lat, 64)

		if err != nil {
			logger.Fatal("Can not parse point because %s", err)
		}

		fl_lon, err := strconv.ParseFloat(str_lon, 64)

		if err != nil {
			logger.Fatal("Can not parse point because %s", err)
		}

		*lat = fl_lat
		*lon = fl_lon
	}

	var db *database.SQLiteDatabase

	if *engine == "spatialite" {

		d, err := database.NewDBWithDriver(*engine, *dsn)

		if err != nil {
			logger.Fatal("Failed to create spatialite database, because %s", err)
		}

		db = d
	}

	/*
		appcache_opts, err := app.DefaultApplicationCacheOptions()

		if err != nil {
			logger.Fatal("Failed to creation application cache options, because %s", err)
		}

		appcache_opts.IndexMode = *mode
		appcache_opts.IndexPaths = flag.Args()

		appcache_opts.FailoverCache = *failover_cache_enable

		appcache_opts.LRUCache = *lru_cache_enable
		appcache_opts.LRUCacheSize = *lru_cache_size
		appcache_opts.LRUCacheTriggerSize = *lru_cache_trigger

		appcache_opts.SourceCache = *source_cache_enable
		appcache_opts.SourceCacheRoot = *source_cache_root

		appcache, err := app.ApplicationCache(appcache_opts)
	*/

	appcache, err := cache.NewSpatialiteCache(db)

	if err != nil {
		logger.Fatal("Failed to creation application cache, because %s", err)
	}

	var appindex pip.Index
	var appindex_err error

	switch *engine {
	case "rtree":
		appindex, appindex_err = pip.NewRTreeIndex(appcache)
	case "spatialite":
		appindex, appindex_err = pip.NewSpatialiteIndex(db, appcache)
	default:
		appindex_err = errors.New("Invalid engine")
	}

	if appindex_err != nil {
		logger.Fatal("failed to create index because %s", appindex_err)
	}

	/*
		appindex, err := app.ApplicationIndex(appcache)

		if err != nil {
			logger.Fatal("failed to create index because %s", err)
		}
	*/

	if *mode != "spatialite" {
		indexer_opts, err := app.DefaultApplicationIndexerOptions()

		if err != nil {
			logger.Fatal("failed to create indexer options because %s", err)
		}

		indexer_opts.IndexMode = *mode

		indexer, err := app.NewApplicationIndexer(appindex, indexer_opts)

		err = indexer.IndexPaths(flag.Args())

		if err != nil {
			logger.Fatal("failed to index paths because %s", err)
		}
	}

	logger.Status("cache size: %d evictions: %d", appcache.Size(), appcache.Evictions())

	f, err := filter.NewSPRFilter()

	if err != nil {
		logger.Fatal("failed to create filter because %s", err)
	}

	if *interactive {

		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {

			input := scanner.Text()
			logger.Status(input)

			parts := strings.Split(input, ",")

			if len(parts) != 2 {
				logger.Warning("Invalid input")
				continue
			}

			str_lat := strings.Trim(parts[0], " ")
			str_lon := strings.Trim(parts[1], " ")

			lat, err := strconv.ParseFloat(str_lat, 64)

			if err != nil {
				logger.Warning("Invalid latitude, %s", err)
				continue
			}

			lon, err := strconv.ParseFloat(str_lon, 64)

			if err != nil {
				logger.Warning("Invalid longitude, %s", err)
				continue
			}

			err = PIPLatLon(appindex, lat, lon, f, logger)

			if err != nil {
				logger.Warning("Failed to PIP, %s", err)
				continue
			}
		}

	} else {

		err = PIPLatLon(appindex, *lat, *lon, f, logger)

		if err != nil {
			logger.Fatal("Failed to PIP, %s", err)
		}
	}

	os.Exit(0)
}
