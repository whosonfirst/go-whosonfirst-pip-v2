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
	_ "log"
	"strings"
	"sync"
)

func NewApplicationIndexer(fl *flag.FlagSet, appindex index.Index, appextras *database.SQLiteDatabase) (*wof_index.Indexer, error) {

	mode, _ := flags.StringVar(fl, "mode")
	is_wof, _ := flags.BoolVar(fl, "is-wof")

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
