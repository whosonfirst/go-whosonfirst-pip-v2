package main

import (
	"context"
	"flag"
	"fmt"
	wof_index "github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"github.com/whosonfirst/go-whosonfirst-sqlite/index"
	"github.com/whosonfirst/go-whosonfirst-sqlite/tables"
	"io"
	"os"
	"runtime"
	"strings"
)

// THIS IS A TOTAL HACK UNTIL WE CAN SORT THINGS OUT IN
// go-whosonfirst-index... (20180206/thisisaaronland)

type Closer struct {
	fh io.Reader
}

func (c Closer) Read(b []byte) (int, error) {
	return c.fh.Read(b)
}

func (c Closer) Close() error {
	return nil
}

func main() {

	valid_modes := strings.Join(wof_index.Modes(), ",")
	desc_modes := fmt.Sprintf("The mode to use importing data. Valid modes are: %s.", valid_modes)

	mode := flag.String("mode", "files", desc_modes)

	dsn := flag.String("dsn", ":memory:", "")
	driver := flag.String("driver", "sqlite3", "")

	live_hard := flag.Bool("live-hard-die-fast", false, "Enable various performance-related pragmas at the expense of possible (unlikely) database corruption")
	timings := flag.Bool("timings", false, "Display timings during and after indexing")

	var procs = flag.Int("processes", (runtime.NumCPU() * 2), "The number of concurrent processes to index data with")

	flag.Parse()

	runtime.GOMAXPROCS(*procs)

	logger := log.SimpleWOFLogger()

	stdout := io.Writer(os.Stdout)
	logger.AddLogger(stdout, "status")

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

	ex, err := tables.NewExampleTableWithDatabase(db)

	if err != nil {
		logger.Fatal("failed to create 'geojson' table because %s", err)
	}

	to_index = append(to_index, ex)

	cb := func(ctx context.Context, fh io.Reader, args ...interface{}) (interface{}, error) {

		return fh, nil

		// HACK - see above
		// closer := Closer{fh}
		// return Closer
	}

	idx, err := index.NewSQLiteIndexer(db, to_index, cb)

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
