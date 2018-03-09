package app

import (
	"flag"
	wof_index "github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/flags"
	"github.com/whosonfirst/go-whosonfirst-pip/index"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"runtime/debug"
	"time"
)

type PIPApplication struct {
	mode    string
	Index   index.Index
	Cache   cache.Cache
	Extras  *database.SQLiteDatabase
	Indexer *wof_index.Indexer
	Logger  *log.WOFLogger
}

func NewPIPApplication(fl *flag.FlagSet) (*PIPApplication, error) {

	logger, err := NewApplicationLogger(fl)

	if err != nil {
		return nil, err
	}

	appcache, err := NewApplicationCache(fl)

	if err != nil {
		return nil, err
	}

	appindex, err := NewApplicationIndex(fl, appcache)

	if err != nil {
		return nil, err
	}

	appextras, err := NewApplicationExtras(fl)

	if err != nil {
		return nil, err
	}

	indexer, err := NewApplicationIndexer(fl, appindex, appextras)

	if err != nil {
		return nil, err
	}

	mode, _ := flags.StringVar(fl, "mode")

	p := PIPApplication{
		mode:    mode,
		Cache:   appcache,
		Index:   appindex,
		Extras:  appextras,
		Indexer: indexer,
		Logger:  logger,
	}

	return &p, nil
}

func (p *PIPApplication) Close() error {

	p.Cache.Close()
	p.Index.Close()

	if p.Extras != nil {
		p.Extras.Close()
	}

	return nil
}

func (p *PIPApplication) IndexPaths(paths []string) error {

	if p.mode != "spatialite" {

		go func() {

			// TO DO: put this somewhere so that it can be triggered by signal(s)
			// to reindex everything in bulk or incrementally

			t1 := time.Now()

			err := p.Indexer.IndexPaths(paths)

			if err != nil {
				p.Logger.Fatal("failed to index paths because %s", err)
			}

			t2 := time.Since(t1)

			p.Logger.Status("finished indexing in %v", t2)
			debug.FreeOSMemory()
		}()

		// set up some basic monitoring and feedback stuff

		go func() {

			c := time.Tick(1 * time.Second)

			for _ = range c {

				if !p.Indexer.IsIndexing() {
					continue
				}

				p.Logger.Status("indexing %d records indexed", p.Indexer.Indexed)
			}
		}()
	}

	return nil
}
