package app

import (
	"context"
	"database/sql"
	"flag"
	"github.com/tidwall/gjson"
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
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
)

func NewApplicationIndexer(fl *flag.FlagSet, appindex index.Index) (*wof_index.Indexer, error) {

	pip_cache, _ := flags.StringVar(fl, "cache")
	mode, _ := flags.StringVar(fl, "mode")
	is_wof, _ := flags.BoolVar(fl, "is-wof")

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

	// extras...
	// FIX ME...

	include_extras := false
	var extras_dsn string

	enable_extras, _ := flags.BoolVar(fl, "enable_extras")

	if enable_extras {

		index_extras := true

		// we are relying on the fact that all of these things have already
		// been vetted above and that the spatialite DB in fact has a geojson
		// table (20180228/thisisaaronland)

		// the problem with this approach is that we might be using a SQLite
		// database that was *generated* by the cache/sqlite.go code whose Set()
		// method only knows about cache.CacheItem thingies which don't have a
		// full WOF properties hash so things like '?extras=geom:longitude'
		// will always fail... (20180228/thisisaaronland)

		// for example, this:
		// ./bin/wof-pip-server -index spatialite -cache spatialite -spatialite dsn=test3.db -enable-extras
		//
		// where test3.db has previously been created by doing (something like) this:
		// ./bin/wof-pip -index spatialite -cache spatialite -spatialite dsn=test3.db -mode repo /usr/local/data/whosonfirst-data
		//
		// which will have populated the 'geojson' table in 'test3.db' using the cache.Set()
		// method described above, and which will be lacking a full (WOF) properties
		// dictionary
		//
		// possible solutions include:
		//
		// 1. testing for and using a '-extras dsn=foo.db' flag which has the perverse
		//    side-effect of requiring *two* SQLite databases
		// 2. testing the '-spatialite dsn=foo.db' database for a record that contains
		//    something we know will be in the WOF properties hash but is _not_ part of
		//    the SPR interface (geom:latitude for example) and throwing an error if it
		//    is missing
		// 3. changing the name of the table that the sqlite.Cache Get() method uses and
		//    adding a flag (flags) to query the correct table and... I am having trouble
		//    keeping track of it as I write these words
		//
		// (2) plus proper documentation is probably the easiest thing going forward under
		// the assumption that almost no one is going to be creating *fresh* databases and
		// instead just using the databases that WOF itself produces (20180228/thisisaaronland)

		spatialite_dsn, _ := flags.StringVar(fl, "spatialite-dsn")

		if pip_cache == "spatialite" || pip_cache == "sqlite" {

			dsn := spatialite_dsn

			// see above - this is solution (2) which is pretty WOF-specific in that it
			// tests for a geom:latitude property which will probably break things if
			// someone is indexing not-WOF documents but we'll just file that as a
			// known-known for now (20180228/thisisaaronland)

			if dsn != ":memory:" {

				db_test, err := database.NewDB(dsn)

				if err != nil {
					return nil, err
				}

				defer db_test.Close()

				conn, err := db_test.Conn()

				if err != nil {
					return nil, err
				}

				row := conn.QueryRow("SELECT body FROM geojson LIMIT 1")

				var body []byte
				err = row.Scan(&body)

				switch {
				case err == sql.ErrNoRows:
					return nil, err
				case err != nil:
					return nil, err
				default:
					// pass
				}

				geom_lat := gjson.GetBytes(body, "properties.geom:latitude")

				if !geom_lat.Exists() {
					return nil, err
				}

				db_test.Close()

				index_extras = false
				extras_dsn = dsn
			}
		}

		if index_extras {

			dsn := spatialite_dsn

			// MAYBE REVISIT THIS DECISION? (20180228/thisisaaronland)

			if dsn == ":memory:" {

				tmpfile, err := ioutil.TempFile("", "pip-extras")

				if err != nil {
					return nil, err
				}

				tmpfile.Close()
				tmpnam := tmpfile.Name()

				dsn = tmpnam

				cleanup := func() {

					// logger.Status("remove temporary extras database '%s'", tmpnam)

					err := os.Remove(tmpnam)

					if err != nil {
						// logger.Warning("failed to remove %s, because %s", tmpnam, err)
					}
				}

				defer cleanup()

				signal_ch := make(chan os.Signal, 1)
				signal.Notify(signal_ch, os.Interrupt)

				go func() {
					<-signal_ch
					cleanup()
				}()
			}

			extras_dsn = dsn
		}

		// FIX ME
		// indexer_opts.IndexExtras = index_extras
		// indexer_opts.ExtrasDB = extras_dsn
	}

	//

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
