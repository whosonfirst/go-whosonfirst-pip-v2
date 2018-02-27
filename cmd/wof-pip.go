package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	geojson_utils "github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-pip/app"
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	"github.com/whosonfirst/go-whosonfirst-pip/flags"
	"github.com/whosonfirst/go-whosonfirst-pip/index"
	"github.com/whosonfirst/go-whosonfirst-pip/utils"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	// golog "log"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func main() {

	var pip_index = flag.String("index", "rtree", "Valid options are: rtree, spatialite")
	var pip_cache = flag.String("cache", "gocache", "Valid options are: gocache, fs, spatialite")

	var mode = flag.String("mode", "files", "")
	var procs = flag.Int("processes", runtime.NumCPU()*2, "")

	var fs_args flags.KeyValueArgs
	flag.Var(&fs_args, "fs-cache", "(0) or more user-defined '{KEY}={VALUE}' arguments to pass to the fs cache")

	var spatialite_args flags.KeyValueArgs
	flag.Var(&spatialite_args, "spatialite-index", "(0) or more user-defined '{KEY}={VALUE}' arguments to pass to the spatialite database")

	var verbose = flag.Bool("verbose", false, "")

	flag.Parse()

	runtime.GOMAXPROCS(*procs)

	logger := log.SimpleWOFLogger()

	if *verbose {
		stdout := io.Writer(os.Stdout)
		logger.AddLogger(stdout, "info")
	}

	var db *database.SQLiteDatabase

	var appindex index.Index
	var appindex_err error

	var appcache cache.Cache
	var appcache_err error

	logger.Info("index is %s cache is %s", *pip_index, *pip_cache)

	if *pip_index == "spatialite" {

		args := spatialite_args.ToMap()
		dsn, ok := args["dsn"]

		if !ok {
			dsn = ":memory:"
		}

		d, err := database.NewDBWithDriver(*pip_index, dsn)

		if err != nil {
			logger.Fatal("Failed to create spatialite database, because %s", err)
		}

		err = d.LiveHardDieFast()

		if err != nil {
			logger.Fatal("Failed to create spatialite database, because %s", err)
		}

		db = d
	}

	switch *pip_cache {

	case "gocache":

		opts, err := cache.DefaultGoCacheOptions()

		if err != nil {
			appcache_err = err
		} else {
			appcache, appcache_err = cache.NewGoCache(opts)
		}

	case "fs":

		args := fs_args.ToMap()
		root, ok := args["root"]

		if ok {
			appcache, appindex_err = cache.NewFSCache(root)
		} else {
			appcache_err = errors.New("Missing FS cache root")
		}

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

	// note: this is "-mode spatialite" not "-engine spatialite"

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

	fmt.Println("ready to query")

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {

		input := scanner.Text()
		logger.Status("# %s", input)

		parts := strings.Split(input, " ")

		if len(parts) == 0 {
			logger.Warning("Invalid input")
			continue
		}

		var command string

		switch parts[0] {

		case "candidates":
			command = parts[0]
		case "pip":
			command = parts[0]
		case "polyline":
			command = parts[0]
		default:
			logger.Warning("Invalid command")
			continue
		}

		var results interface{}

		if command == "pip" || command == "candidates" {

			str_lat := strings.Trim(parts[1], " ")
			str_lon := strings.Trim(parts[2], " ")

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

			c, err := geojson_utils.NewCoordinateFromLatLons(lat, lon)

			if err != nil {
				logger.Warning("Invalid latitude, longitude, %s", err)
				continue
			}

			if command == "pip" {

				intersects, err := appindex.GetIntersectsByCoord(c, f)

				if err != nil {
					logger.Warning("Unable to get intersects, because %s", err)
					continue
				}

				results = intersects

			} else {

				candidates, err := appindex.GetCandidatesByCoord(c)

				if err != nil {
					logger.Warning("Unable to get candidates, because %s", err)
					continue
				}

				results = candidates
			}

		} else if command == "polyline" {

			poly := parts[1]
			factor := 1.0e5

			if len(parts) > 2 {

				f, err := utils.StringPrecisionToFactor(parts[2])

				if err != nil {
					logger.Warning("Unable to parse precision because %s", err)
					continue
				}

				factor = f
			}

			path, err := utils.DecodePolyline(poly, factor)

			if err != nil {
				logger.Warning("Unable to decode polyline because %s", err)
				continue
			}

			intersects, err := appindex.GetIntersectsByPath(*path, f)

			if err != nil {
				logger.Warning("Unable to get candidates, because %s", err)
				continue
			}

			logger.Info("intersects %v", intersects)
			results = intersects

		} else {
			logger.Warning("Invalid command")
			continue
		}

		body, err := json.Marshal(results)

		if err != nil {
			logger.Warning("Failed to marshal results, because %s", err)
			continue
		}

		fmt.Println(string(body))
	}

	os.Exit(0)
}
