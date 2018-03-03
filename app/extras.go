package app

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-pip/flags"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"io/ioutil"
	"os"
	"os/signal"
)

func NewApplicationExtras(fl *flag.FlagSet) (*database.SQLiteDatabase, error) {

	enable_extras, _ := flags.BoolVar(fl, "enable-extras")
	extras_dsn, _ := flags.StringVar(fl, "extras-dsn")

	if !enable_extras {
		return nil, nil
	}

	if extras_dsn == ":tmpfile:" {

		tmpfile, err := ioutil.TempFile("", "pip-extras")

		if err != nil {
			return nil, err
		}

		tmpfile.Close()
		tmpnam := tmpfile.Name()

		extras_dsn = tmpnam

		cleanup := func() {

			// logger.Status("remove temporary extras database '%s'", tmpnam)

			err := os.Remove(extras_dsn)

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

	db, err := database.NewDB(extras_dsn)

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

	return db, nil
}
