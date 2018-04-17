package main

import (
	"flag"
	"fmt"
	wof_index "github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/index"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"io"
	"os"
	"runtime"
	"strings"
)

func main() {

	valid_modes := strings.Join(wof_index.Modes(), ",")
	desc_modes := fmt.Sprintf("The mode to use importing data. Valid modes are: %s.", valid_modes)

	dsn := flag.String("dsn", ":memory:", "")
	driver := flag.String("driver", "sqlite3", "")

	mode := flag.String("mode", "files", desc_modes)

	all := flag.Bool("all", false, "Index all tables (except the 'search' and 'geometries' tables which you need to specify explicitly)")
	ancestors := flag.Bool("ancestors", false, "Index the 'ancestors' tables")
	concordances := flag.Bool("concordances", false, "Index the 'concordances' tables")
	geojson := flag.Bool("geojson", false, "Index the 'geojson' table")
	geometries := flag.Bool("geometries", false, "Index the 'geometries' table (requires that libspatialite already be installed)")
	names := flag.Bool("names", false, "Index the 'names' table")
	search := flag.Bool("search", false, "Index the 'search' table (using SQLite FTS4 full-text indexer)")
	spr := flag.Bool("spr", false, "Index the 'spr' table")
	live_hard := flag.Bool("live-hard-die-fast", true, "Enable various performance-related pragmas at the expense of possible (unlikely) database corruption")
	timings := flag.Bool("timings", false, "Display timings during and after indexing")
	// liberal := flag.Bool("liberal", false, "Do not trigger errors for records that can not be processed, for whatever reason")
	var procs = flag.Int("processes", (runtime.NumCPU() * 2), "The number of concurrent processes to index data with")

	flag.Parse()

	runtime.GOMAXPROCS(*procs)

	logger := log.SimpleWOFLogger()

	stdout := io.Writer(os.Stdout)
	logger.AddLogger(stdout, "status")

	if *geometries && *driver != "spatialite" {
		logger.Fatal("you asked to index geometries but specified the '%s' driver instead of spatialite", *driver)
	}

	db, err := database.NewDBWithDriver(*driver, *dsn)

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

	if *concordances || *all {

		cn, err := tables.NewConcordancesTableWithDatabase(db)

		if err != nil {
			logger.Fatal("failed to create 'concordances' table because %s", err)
		}

		to_index = append(to_index, cn)
	}

	// see the way we don't check *all here - that's so people who don't have
	// spatialite installed can still use *all (20180122/thisisaaronland)

	if *geometries {

		gm, err := tables.NewGeometriesTableWithDatabase(db)

		if err != nil {
			logger.Fatal("failed to create 'geometries' table because %s", err)
		}

		to_index = append(to_index, gm)
	}

	// see the way we don't check *all here either - that's because this table can be
	// brutally slow to index and should probably really just be a separate database
	// anyway... (20180214/thisisaaronland)

	if *search {

		st, err := tables.NewSearchTableWithDatabase(db)

		if err != nil {
			logger.Fatal("failed to create 'search' table because %s", err)
		}

		to_index = append(to_index, st)
	}

	if len(to_index) == 0 {
		logger.Fatal("You forgot to specify which (any) tables to index")
	}

	idx, err := index.NewDefaultSQLiteFeaturesIndexer(db, to_index)

	if err != nil {
		logger.Fatal("failed to create sqlite indexer because %s", err)
	}

	idx.Timings = *timings
	idx.Logger = logger

	err = idx.IndexPaths(*mode, flag.Args())

	if err != nil {
		logger.Fatal("Failed to index paths in %s mode because: %s", *mode, err)
	}

	os.Exit(0)
}
