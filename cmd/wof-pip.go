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
	"github.com/whosonfirst/go-whosonfirst-pip/flags"
	"github.com/whosonfirst/go-whosonfirst-pip/index"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	// golog "log"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func PIPLatLon(i index.Index, lat float64, lon float64, f filter.Filter, logger *log.WOFLogger) error {

	c, err := utils.NewCoordinateFromLatLons(lat, lon)

	if err != nil {
		return err
	}

	return PIP(i, c, f, logger)
}

func PIP(i index.Index, c geom.Coord, f filter.Filter, logger *log.WOFLogger) error {

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

	/*
		t1 := time.Now()

		cd, err := i.GetCandidatesByCoord(c)

		cd_body, err := json.Marshal(cd)

		t2 := time.Since(t1)
		logger.Status("time to fetch candidates: %v\n", t2)

		if err != nil {
			return err
		}

		fmt.Println(string(cd_body))
	*/

	return nil
}

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
			appindex, appindex_err = index.NewFSCache(root)
		} else {
			appindex_err = errors.New("Missing FS cache root")
		}

	case "sqlite":
		appindex, appindex_err = index.NewSQLiteCache(db)
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

	os.Exit(0)
}
