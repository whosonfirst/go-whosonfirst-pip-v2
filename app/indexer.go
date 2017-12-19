package app

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/geometry"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-index"
	pip "github.com/whosonfirst/go-whosonfirst-pip/index"
	pip_utils "github.com/whosonfirst/go-whosonfirst-pip/utils"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"github.com/whosonfirst/go-whosonfirst-sqlite/tables"
	"io"
	"log"
	"sync"
)

type ApplicationIndexerOptions struct {
	IndexMode         string
	IsWOF             bool
	IncludeDeprecated bool
	IncludeSuperseded bool
	IncludeCeased     bool
	IncludeNotCurrent bool
	IndexExtras       bool
	ExtrasDB          string
}

func DefaultApplicationIndexerOptions() (ApplicationIndexerOptions, error) {

	opts := ApplicationIndexerOptions{
		IndexMode:         "",
		IsWOF:             true,
		IncludeDeprecated: true,
		IncludeSuperseded: true,
		IncludeCeased:     true,
		IncludeNotCurrent: true,
		IndexExtras:       false,
		ExtrasDB:          "",
	}

	return opts, nil
}

func NewApplicationIndexer(appindex pip.Index, opts ApplicationIndexerOptions) (*index.Indexer, error) {

	var wg *sync.WaitGroup
	var mu *sync.Mutex
	var gt sqlite.Table

	if opts.IndexExtras {

		db, err := database.NewDB(opts.ExtrasDB)

		if err != nil {
			return nil, err
		}

		defer db.Close()

		err = db.LiveHardDieFast()	// otherwise indexing will be brutally slow with large datasets

		if err != nil {
			return nil, err
		}

		// see also:
		// https://github.com/whosonfirst/go-whosonfirst-pip-v2/issues/19
		
		gt, err = tables.NewGeoJSONTableWithDatabase(db)

		if err != nil {
			return nil, err
		}

		wg = new(sync.WaitGroup)
		mu = new(sync.Mutex)
	}

	cb := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

		var f geojson.Feature

		if opts.IsWOF {

			ok, err := pip_utils.IsValidRecord(fh, ctx)

			if err != nil {
				return err
			}

			if !ok {
				return err
			}

			tmp, err := feature.LoadWOFFeatureFromReader(fh)

			if err != nil {
				return err
			}

			if !opts.IncludeNotCurrent {

				fl, err := whosonfirst.IsCurrent(tmp)

				if err != nil {
					return err
				}

				if fl.IsTrue() && fl.IsKnown() {
					return nil
				}
			}

			if !opts.IncludeDeprecated {

				fl, err := whosonfirst.IsDeprecated(tmp)

				if err != nil {
					return err
				}

				if fl.IsTrue() && fl.IsKnown() {
					return nil
				}
			}

			if !opts.IncludeCeased {

				fl, err := whosonfirst.IsCeased(tmp)

				if err != nil {
					return err
				}

				if fl.IsTrue() && fl.IsKnown() {
					return nil
				}
			}

			if !opts.IncludeSuperseded {

				fl, err := whosonfirst.IsSuperseded(tmp)

				if err != nil {
					return err
				}

				if fl.IsTrue() && fl.IsKnown() {
					return nil
				}
			}

			f = tmp

		} else {

			tmp, err := feature.LoadFeatureFromReader(fh)

			if err != nil {
				return err
			}

			f = tmp
		}

		geom_type := geometry.Type(f)

		if geom_type == "Point" {
			return nil
		}

		err := appindex.IndexFeature(f)

		if err != nil {
			return err
		}

		// see also: http/intersects.go (20171217/thisisaaronland)

		// notice the way errors indexing things in SQLite do not trigger
		// an error signal - maybe we want to do that? maybe not...?
		// (20171218/thisisaaronland)

		if opts.IndexExtras {

			wg.Add(1)

			go func(f geojson.Feature, wg *sync.WaitGroup) error {

				defer wg.Done()

				db, err := database.NewDB(opts.ExtrasDB)

				if err != nil {
					log.Println(err)
					return err
				}

				defer db.Close()

				mu.Lock()

				err = gt.IndexFeature(db, f)

				mu.Unlock()

				if err != nil {
					// log.Println("FAILED TO INDEX", err) // something
				}

				return err

			}(f, wg)
		}

		return nil
	}

	idx, err := index.NewIndexer(opts.IndexMode, cb)

	if opts.IndexExtras {
		wg.Wait()
	}

	return idx, err
}
