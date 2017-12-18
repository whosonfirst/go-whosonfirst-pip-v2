package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-index/utils"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"github.com/whosonfirst/go-whosonfirst-sqlite/tables"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func main() {

	valid_modes := strings.Join(index.Modes(), ",")
	desc_modes := fmt.Sprintf("The mode to use importing data. Valid modes are: %s.", valid_modes)

	dsn := flag.String("dsn", ":memory:", "")
	mode := flag.String("mode", "files", desc_modes)

	all := flag.Bool("all", false, "Index all tables")
	ancestors := flag.Bool("ancestors", false, "Index the 'ancestors' tables")
	geojson := flag.Bool("geojson", false, "Index the 'geojson' table")
	names := flag.Bool("names", false, "Index the 'names' table")
	spr := flag.Bool("spr", false, "Index the 'spr' table")
	live_hard := flag.Bool("live-hard-die-fast", false, "Enable various performance-related pragmas at the expense of possible (unlikely) database corruption")
	timings := flag.Bool("timings", false, "Display timings during and after indexing")
	var procs = flag.Int("processes", (runtime.NumCPU() * 2), "The number of concurrent processes to index data with")

	flag.Parse()

	runtime.GOMAXPROCS(*procs)

	logger := log.SimpleWOFLogger()

	db, err := database.NewDB(*dsn)

	if err != nil {
		logger.Fatal("unable to create database (%s) because %s", *dsn, err)
	}

	defer db.Close()

	if *live_hard {

		err = db.LiveHardDieFast()

		if err != nil {
			logger.Fatal("Unable to live hard and die fast so just dying fast instead, because %s", err)
		}
	}

	to_index := make([]sqlite.Table, 0)

	if *geojson || *all {

		gt, err := tables.NewGeoJSONTableWithDatabase(db)

		if err != nil {
			logger.Fatal("failed to create 'geojson' table because %s", err)
		}

		to_index = append(to_index, gt)
	}

	if *spr || *all {

		st, err := tables.NewSPRTableWithDatabase(db)

		if err != nil {
			logger.Fatal("failed to create 'spr' table because %s", err)
		}

		to_index = append(to_index, st)
	}

	if *names || *all {

		nm, err := tables.NewNamesTableWithDatabase(db)

		if err != nil {
			logger.Fatal("failed to create 'names' table because %s", err)
		}

		to_index = append(to_index, nm)
	}

	if *ancestors || *all {

		an, err := tables.NewAncestorsTableWithDatabase(db)

		if err != nil {
			logger.Fatal("failed to create 'ancestors' table because %s", err)
		}

		to_index = append(to_index, an)
	}

	if len(to_index) == 0 {
		logger.Fatal("You forgot to specify which (any) tables to index")
	}

	table_timings := make(map[string]time.Duration)
	mu := new(sync.RWMutex)

	cb := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

		path, err := index.PathForContext(ctx)

		if err != nil {
			return err
		}

		ok, err := utils.IsPrincipalWOFRecord(fh, ctx)

		if err != nil {
			return err
		}

		if !ok {
			return nil
		}

		f, err := feature.LoadWOFFeatureFromReader(fh)

		if err != nil {
			logger.Warning("failed to load feature (%s) because %s", path, err)
			return err
		}

		db.Lock()

		defer db.Unlock()

		for _, t := range to_index {

			t1 := time.Now()

			err = t.IndexFeature(db, f)

			if err != nil {
				logger.Warning("failed to index feature (%s) in '%s' table because %s", path, t.Name(), err)
				return err
			}

			t2 := time.Since(t1)

			n := t.Name()

			mu.Lock()

			_, ok := table_timings[n]

			if ok {
				table_timings[n] += t2
			} else {
				table_timings[n] = t2
			}

			mu.Unlock()
		}

		return nil
	}

	indexer, err := index.NewIndexer(*mode, cb)

	if err != nil {
		logger.Fatal("Failed to create new indexer because: %s", err)
	}

	done_ch := make(chan bool)
	t1 := time.Now()

	show_timings := func() {

		t2 := time.Since(t1)

		i := atomic.LoadInt64(&indexer.Indexed) // please just make this part of go-whosonfirst-index

		mu.RLock()
		defer mu.RUnlock()

		for t, d := range table_timings {
			logger.Status("time to index %s (%d) : %v", t, i, d)
		}

		logger.Status("time to index all (%d) : %v", i, t2)
	}

	if *timings {

		go func() {

			for {

				select {
				case <-done_ch:
					return
				case <-time.After(1 * time.Minute):
					show_timings()
				}
			}
		}()

	}

	err = indexer.IndexPaths(flag.Args())

	if err != nil {
		logger.Fatal("Failed to index paths in %s mode because: %s", *mode, err)
	}

	done_ch <- true
	show_timings()

	os.Exit(0)
}
