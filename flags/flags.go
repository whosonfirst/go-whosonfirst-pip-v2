package flags

import (
	"errors"
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-index"
	_ "log"
	"os"
	"strings"
)

func Parse(fl *flag.FlagSet) {

	args := os.Args[1:]

	if len(args) > 0 && args[0] == "-h" {
		fl.Usage()
		os.Exit(0)
	}

	fl.Parse(args)
}

func Lookup(fl *flag.FlagSet, k string) (interface{}, error) {

	v := fl.Lookup(k)

	if v == nil {
		msg := fmt.Sprintf("Unknown flag '%s'", k)
		return nil, errors.New(msg)
	}

	// Go is weird...
	return v.Value.(flag.Getter).Get(), nil
}

func StringVar(fl *flag.FlagSet, k string) (string, error) {

	i, err := Lookup(fl, k)

	if err != nil {
		return "", err
	}

	return i.(string), nil
}

func IntVar(fl *flag.FlagSet, k string) (int, error) {

	i, err := Lookup(fl, k)

	if err != nil {
		return 0, err
	}

	return i.(int), nil
}

func BoolVar(fl *flag.FlagSet, k string) (bool, error) {

	i, err := Lookup(fl, k)

	if err != nil {
		return false, err
	}

	return i.(bool), nil
}

func NewFlagSet(name string) *flag.FlagSet {

	fs := flag.NewFlagSet(name, flag.ExitOnError)

	fs.Usage = func() {
		fs.PrintDefaults()
	}

	return fs
}

func CommonFlags() (*flag.FlagSet, error) {

	fs := NewFlagSet("common")

	fs.String("index", "rtree", "Valid options are: rtree, spatialite.")
	fs.String("cache", "gocache", "Valid options are: gocache, fs, spatialite, sqlite. Note that the spatalite option is just a convenience to mirror the '-index spatialite' option.")

	valid_modes := strings.Join(index.Modes(), ", ")
	desc_modes := fmt.Sprintf("Valid modes are: %s.", valid_modes)

	fs.String("mode", "files", desc_modes)

	fs.String("spatialite-dsn", ":memory:", "A valid SQLite DSN for the '-cache spatialite/sqlite' or '-index spatialite' option. As of this writing for the '-index' and '-cache' options share the same '-spatailite' DSN.")
	fs.String("fs-path", "", "The root directory to look for features if '-cache fs'.")

	fs.Bool("is-wof", true, "Input data is WOF-flavoured GeoJSON.")

	// this is invoked/used in app/indexer.go but for the life of me I can't
	// figure out how to make the code in flags/exclude.go implement the
	// correct inferface wah wah so that flag.Lookup("exclude").Value returns
	// something we can loop over... so instead we just strings.Split() on
	// flag.Lookup("exclude").String() which is dumb but works...
	// (20180301/thisisaaronland)

	var exclude Exclude
	fs.Var(&exclude, "exclude", "Exclude (WOF) records based on their existential flags. Valid options are: ceased, deprecated, not-current, superseded.")

	fs.Bool("verbose", false, "Be chatty.")

	fs.Bool("www", false, "This flag is DEPRECATED. Please use the '-enable-www' flag instead.")
	fs.Bool("polylines", false, "This flag is DEPRECATED. Please use the '-enable-polylines' flag instead.")
	fs.Bool("candidates", false, "This flag is DEPRECATED. Please use the '-enable-candidates' flag instead.")
	fs.Bool("allow-geojson", false, "This flag is DEPRECATED. Please use the '-enable-geojson' flag instead.")
	fs.String("mapzen-api-key", "", "This flag is DEPRECATED. Please use the '-www-api-key' flag instead.")

	fs.String("www-local", "", "This flag is DEPRECATED and doesn't do anything anymore.")
	fs.String("www-local-root", "", "This flag is DEPRECATED and doesn't do anything anymore.")

	fs.String("source-cache-root", "", "This flag is DEPRECATED and doesn't do anything anymore. Please use the '-cache fs' and '-fs-path {PATH}' flags instead.")

	fs.Bool("cache-all", false, "This flag is DEPRECATED and doesn't do anything anymore.")
	fs.String("failover-cache", "", "This flag is DEPRECATED and doesn't do anything anymore.")
	fs.Int("lru-cache-size", 0, "This flag is DEPRECATED and doesn't do anything anymore.")
	fs.Int("lru-cache-trigger", 0, "This flag is DEPRECATED and doesn't do anything anymore.")

	fs.Int("processes", 0, "This flag is DEPRECATED and doesn't do anything anymore.")

	return fs, nil
}
