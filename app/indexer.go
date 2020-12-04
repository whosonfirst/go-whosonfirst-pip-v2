package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/geometry"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	wof_index "github.com/whosonfirst/go-whosonfirst-index"
	_ "github.com/whosonfirst/go-whosonfirst-index/fs"
	"github.com/whosonfirst/go-whosonfirst-pip-v2/flags"
	"github.com/whosonfirst/go-whosonfirst-pip-v2/index"
	"github.com/whosonfirst/go-whosonfirst-pip-v2/utils"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"github.com/whosonfirst/warning"
	"io"
	"log"
	"strings"
	"sync"
)

func NewApplicationIndexer(fl *flag.FlagSet, appindex index.Index, appextras *database.SQLiteDatabase) (*wof_index.Indexer, error) {

	mode, _ := flags.StringVar(fl, "mode")
	is_wof, _ := flags.BoolVar(fl, "is-wof")

	// something something something, just keep going if indexing
	// a given record fails (20190919/thisisaaronland)
	// strict, _ := flags.BoolVar(fl, "strict")

	index_extras := false

	if appextras != nil {

		if mode != "spatialite" {
			index_extras = true
		}
	}

	include_deprecated := true
	include_superseded := true
	include_ceased := true
	include_notcurrent := true

	exclude_fl := fl.Lookup("exclude")

	if exclude_fl != nil {

		// ugh... Go - why do I have to do this... I am willing
		// to believe I am "doing it wrong" (obviously) but for
		// the life of me I can't figure out how to do it "right"
		// (20180301/thisisaaronland)

		exclude := strings.Split(exclude_fl.Value.String(), " ")

		for _, e := range exclude {

			switch e {
			case "deprecated":
				include_deprecated = false
			case "ceased":
				include_ceased = false
			case "superseded":
				include_superseded = false
			case "not-current":
				include_notcurrent = false
			default:
				continue
			}
		}
	}

	var wg *sync.WaitGroup
	var mu *sync.Mutex
	var gt sqlite.Table

	if index_extras {

		t, err := tables.NewGeoJSONTable()

		if err != nil {
			return nil, err
		}

		gt = t
		wg = new(sync.WaitGroup)
		mu = new(sync.Mutex)
	}

	cb := func(ctx context.Context, fh io.Reader, args ...interface{}) error {

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

				// it's still not clear (to me) what the expected or desired
				// behaviour is / in this instance we might be issuing a warning
				// from the geojson-v2 package because a feature might have a
				// placetype defined outside of "core" (in the go-whosonfirst-placetypes)
				// package but that shouldn't necessarily trigger a fatal error
				// (20180405/thisisaaronland)

				if !warning.IsWarning(err) {
					return err
				}

				log.Printf("Feature ID %s triggered the following warning: %s\n", tmp.Id(), err)
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

			tmp, err := feature.LoadGeoJSONFeatureFromReader(fh)

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

			// something something something wrapping errors in Go 1.13
			// something something something waiting to see if the GOPROXY is
			// disabled by default in Go > 1.13 (20190919/thisisaaronland)

			msg := fmt.Sprintf("Failed to index %s (%s), %s", f.Id(), f.Name(), err)
			return errors.New(msg)
		}

		// see also: http/intersects.go (20171217/thisisaaronland)

		// notice the way errors indexing things in SQLite do not trigger
		// an error signal - maybe we want to do that? maybe not...?
		// (20171218/thisisaaronland)

		if index_extras {

			wg.Add(1)

			go func(f geojson.Feature, wg *sync.WaitGroup) error {

				defer wg.Done()

				mu.Lock()

				err = gt.IndexRecord(appextras, f)

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

	if index_extras {
		wg.Wait()
	}

	return idx, err
}
