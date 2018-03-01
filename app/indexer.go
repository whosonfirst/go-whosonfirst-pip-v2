package app

import (
	"context"
	"flag"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/geometry"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	wof_index "github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-pip/flags"
	"github.com/whosonfirst/go-whosonfirst-pip/index"
	"github.com/whosonfirst/go-whosonfirst-pip/utils"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
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

func NewApplicationIndexer(fl *flag.FlagSet, appindex index.Index) (*wof_index.Indexer, error) {

	mode, _ := flags.StringVar(fl, "mode")
	is_wof, _ := flags.BoolVar(fl, "is-wof")

	include_deprecated := true
	include_superseded := true
	include_ceased := true
	include_notcurrent := true

	// FIX ME...

	include_extras := false
	extras_dsn := ""

	var wg *sync.WaitGroup
	var mu *sync.Mutex
	var gt sqlite.Table

	if include_extras {

		db, err := database.NewDB(extras_dsn)

		if err != nil {
			return nil, err
		}

		defer db.Close()

		err = db.LiveHardDieFast() // otherwise indexing will be brutally slow with large datasets

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

		if is_wof {

			ok, err := utils.IsValidRecord(fh, ctx)

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

			if !include_notcurrent {

				fl, err := whosonfirst.IsCurrent(tmp)

				if err != nil {
					return err
				}

				if fl.IsTrue() && fl.IsKnown() {
					return nil
				}
			}

			if !include_deprecated {

				fl, err := whosonfirst.IsDeprecated(tmp)

				if err != nil {
					return err
				}

				if fl.IsTrue() && fl.IsKnown() {
					return nil
				}
			}

			if !include_ceased {

				fl, err := whosonfirst.IsCeased(tmp)

				if err != nil {
					return err
				}

				if fl.IsTrue() && fl.IsKnown() {
					return nil
				}
			}

			if !include_superseded {

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

		if include_extras {

			wg.Add(1)

			go func(f geojson.Feature, wg *sync.WaitGroup) error {

				defer wg.Done()

				db, err := database.NewDB(extras_dsn)

				if err != nil {
					log.Println(err)
					return err
				}

				defer db.Close()

				mu.Lock()

				err = gt.IndexRecord(db, f)

				mu.Unlock()

				if err != nil {
					// log.Println("FAILED TO INDEX", err) // something
				}

				return err

			}(f, wg)
		}

		return nil
	}

	idx, err := wof_index.NewIndexer(mode, cb)

	if include_extras {
		wg.Wait()
	}

	return idx, err
}
