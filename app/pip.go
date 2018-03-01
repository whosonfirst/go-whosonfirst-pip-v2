package app

import (
	"flag"
	wof_index "github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/flags"
	"github.com/whosonfirst/go-whosonfirst-pip/index"
	"log"
	"runtime/debug"
	"time"
)

type PIPApplication struct {
	mode    string
	Index   index.Index
	Cache   cache.Cache
	Indexer *wof_index.Indexer
}

func NewPIPApplication(fl *flag.FlagSet) (*PIPApplication, error) {

	mode, _ := flags.StringVar(fl, "mode")

	appcache, err := NewApplicationCache(fl)

	if err != nil {
		return nil, err
	}

	appindex, err := NewApplicationIndex(fl, appcache)

	if err != nil {
		return nil, err
	}

	indexer, err := NewApplicationIndexer(fl, appindex)

	if err != nil {
		return nil, err
	}

	p := PIPApplication{
		mode:    mode,
		Cache:   appcache,
		Index:   appindex,
		Indexer: indexer,
	}

	return &p, nil
}

func (p *PIPApplication) IndexPaths(paths []string) error {

	if p.mode != "spatialite" {

		go func() {

			// TO DO: put this somewhere so that it can be triggered by signal(s)
			// to reindex everything in bulk or incrementally

			t1 := time.Now()

			err := p.Indexer.IndexPaths(paths)

			if err != nil {
				log.Fatal(err)
				// logger.Fatal("failed to index paths because %s", err)
			}

			t2 := time.Since(t1)

			log.Println(t2)
			//logger.Status("finished indexing in %v", t2)
			debug.FreeOSMemory()
		}()

		// set up some basic monitoring and feedback stuff

		go func() {

			c := time.Tick(1 * time.Second)

			for _ = range c {

				if !p.Indexer.IsIndexing() {
					continue
				}

				// logger.Status("indexing %d records indexed", indexer.Indexed)
			}
		}()
	}

	return nil
}
