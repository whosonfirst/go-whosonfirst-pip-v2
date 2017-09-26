package main

/*

// montreal and montreal
./bin/wof-pip -mode files -latitude 45.593352 -longitude -73.513992 /usr/local/data/whosonfirst-data/data/101/736/545/101736545.geojson
2017/08/04 11:20:24 1

// montreal and the river
./bin/wof-pip -mode files -latitude 45.557093 -longitude -73.513641 /usr/local/data/whosonfirst-data/data/101/736/545/101736545.geojson
2017/08/04 11:20:34 0

// montreal and the westmount
./bin/wof-pip -mode files -latitude 45.486373  -longitude -73.598442 /usr/local/data/whosonfirst-data/data/101/736/545/101736545.geojson
2017/08/04 11:36:43 0

// all the admin data and montreal
./bin/wof-pip -mode repo -cache lru -cache-size 0 -latitude 45.593352 -longitude -73.513992 /usr/local/data/whosonfirst-data/
2017/08/09 16:19:42 time to count 492135 records: 38.539210688s
2017/08/09 16:29:48 time to index records 10m5.612997633s
2017/08/09 16:29:48 time to count 13 records: 20.800389ms

./bin/wof-pip -mode repo -source-cache -latitude 45.593352 -longitude -73.513992 /usr/local/data/whosonfirst-data/
2017/08/09 17:34:40 time to index records 16m9.86740636s
2017/08/09 17:34:44 time to count 13 records: 3.978013837s

./bin/wof-pip -mode repo -failover-cache -lru-cache-size 1024 -lru-cache-trigger 2000 -interactive /usr/local/data/whosonfirst-data/
13:09:03.877576 [wof-pip][STATUS] time to index records 12m29.559567659s
13:09:03.877598 [wof-pip][STATUS] cache size: 1024 evictions: 12202
45.557093,-73.513641
13:19:43.652334 [wof-pip][STATUS] time to count 6 records: 2.642539604s

*/

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/skelterjohn/geom"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-pip/app"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	pip "github.com/whosonfirst/go-whosonfirst-pip/index"
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

	var lat = flag.Float64("latitude", 0.0, "")
	var lon = flag.Float64("longitude", 0.0, "")
	var point = flag.String("point", "", "")

	var mode = flag.String("mode", "files", "")
	var procs = flag.Int("processes", runtime.NumCPU()*2, "")

	var source_cache_enable = flag.Bool("source-cache", false, "")
	var source_cache_root = flag.String("source-cache-data-root", "/usr/local/data", "")

	var lru_cache_enable = flag.Bool("lru-cache", false, "")
	var lru_cache_size = flag.Int("lru-cache-size", 0, "")
	var lru_cache_trigger = flag.Int("lru-cache-trigger", 0, "")

	var failover_cache_enable = flag.Bool("failover-cache", false, "")

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

	if err != nil {
		logger.Fatal("Failed to creation application cache, because %s", err)
	}

	appindex, err := app.ApplicationIndex(appcache)

	if err != nil {
		logger.Fatal("failed to create index because %s", err)
	}

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
