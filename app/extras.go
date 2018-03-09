package app

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-pip/flags"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
)

func NewApplicationExtras(fl *flag.FlagSet) (*database.SQLiteDatabase, error) {

	enable_extras, _ := flags.BoolVar(fl, "enable-extras")
	extras_dsn, _ := flags.StringVar(fl, "extras-dsn")

	if !enable_extras {
		return nil, nil
	}

	var db *database.SQLiteDatabase

	if extras_dsn == ":tmpfile:" {

		tmpfile, err := ioutil.TempFile("", "pip-extras")

		if err != nil {
			return nil, err
		}

		tmpfile.Close()
		tmpnam := tmpfile.Name()

		extras_dsn = tmpnam

		cleanup := func() {

			err := db.Close()

			if err != nil {
				log.Printf("Failed to close extras database (%s) because %s\n", extras_dsn, err)
				return
			}

			err = os.Remove(extras_dsn)

			if err != nil {
				log.Printf("Failed to close extras database (%s) because %s\n", extras_dsn, err)
				return
			}
		}

		signal_ch := make(chan os.Signal, 1)
		signal.Notify(signal_ch, os.Interrupt)

		go func() {
			<-signal_ch
			cleanup()
		}()
	}

	var err error

	db, err = database.NewDB(extras_dsn)

	if err != nil {
		return nil, err
	}

	err = db.LiveHardDieFast() // otherwise indexing will be brutally slow with large datasets

	if err != nil {
		return nil, err
	}

	// see also:
	// https://github.com/whosonfirst/go-whosonfirst-pip-v2/issues/19

	_, err = tables.NewGeoJSONTableWithDatabase(db)

	if err != nil {
		return nil, err
	}

	db.Close()

	db, err = database.NewDB(extras_dsn)

	if err != nil {
		return nil, err
	}

	return db, nil
}
