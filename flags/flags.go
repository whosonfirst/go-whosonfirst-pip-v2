package flags

import (
	"errors"
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-index"
	"log"
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

	err := Validate(fl)

	if err != nil {
		log.Fatal(err)
	}
}

func Validate(fs *flag.FlagSet) error {

	strict, err := BoolVar(fs, "strict")

	if err != nil {
		return err
	}

	mode, err := StringVar(fs, "mode")

	if err != nil {
		return err
	}

	pip_index, err := StringVar(fs, "index")

	if err != nil {
		return err
	}

	pip_cache, err := StringVar(fs, "cache")

	if err != nil {
		return err
	}

	spatialite_dsn, err := StringVar(fs, "spatialite-dsn")

	if err != nil {
		return err
	}

	if mode == "spatialite" {

		if pip_index != "spatialite" {
			return errors.New("-mode is spatialite but -index is not")
		}

		if pip_cache != "sqlite" && pip_cache != "spatialite" {
			return errors.New("-mode is spatialite but -cache is neither 'sqlite' or 'spatialite'")
		}

		if spatialite_dsn == "" || spatialite_dsn == ":memory:" {
			return errors.New("-spatialite-dsn needs to be an actual file on disk")
		}
	}

	deprecated_bool := map[string]string{
		"allow-geojson": "enable-geojson",
		"candidates":    "enable-candidates",
		"polylines":     "enable-polylines",
		"www":           "enable-www",
	}

	for old, new := range deprecated_bool {

		value, err := BoolVar(fs, old)

		if err != nil {
			return err
		}

		if value {

			warning := fmt.Sprintf("deprecated flag -%s used so helpfully assigning -%s flag\n", old, new)

			if strict {
				warning = fmt.Sprintf("deprecated flag -%s used with -strict flag enabled", old)
				return errors.New(warning)
			}

			log.Printf("[WARNING] %s\n", warning)
			fs.Set(new, fmt.Sprintf("%s", value))
		}
	}

	deprecated_string := map[string]string{
		"mapzen-api-key": "www-api-key",
	}

	for old, new := range deprecated_string {

		value, err := StringVar(fs, old)

		if err != nil {
			return err
		}

		if value != "" {

			warning := fmt.Sprintf("deprecated flag -%s used so helpfully assigning -%s flag\n", old, new)

			if strict {
				warning := fmt.Sprintf("deprecated flag -%s used with -strict flag enabled", old)
				return errors.New(warning)
			}

			log.Printf("[WARNING] %s\n", warning)
			fs.Set(new, value)
		}
	}

	invalid_string := []string{
		"www-local",
		"www-local-root",
		"source-cache-root",
	}

	for _, old := range invalid_string {

		value, err := StringVar(fs, old)

		if err != nil {
			return err
		}

		if value != "" {

			warning := fmt.Sprintf("deprecated flag -%s used but it has no meaning anymore", old)

			if strict {
				return errors.New(warning)
			}

			log.Printf("[WARNING] %s\n", warning)
		}
	}

	invalid_bool := []string{
		"cache-all",
	}

	for _, old := range invalid_bool {

		value, err := BoolVar(fs, old)

		if err != nil {
			return err
		}

		if value {

			warning := fmt.Sprintf("deprecated flag -%s used but it has no meaning anymore", old)

			if strict {
				return errors.New(warning)
			}

			log.Printf("[WARNING] %s\n", warning)
		}
	}

	invalid_int := []string{
		"lru-cache-size",
		"lru-cache-trigger",
		"processes",
	}

	for _, old := range invalid_int {

		value, err := IntVar(fs, old)

		if err != nil {
			return err
		}

		if value != 0 {

			warning := fmt.Sprintf("deprecated flag -%s used but it has no meaning anymore", old)

			if strict {
				return errors.New(warning)
			}

			log.Printf("[WARNING] %s\n", warning)
		}
	}

	enable_www, err := BoolVar(fs, "enable-www")

	if err != nil {
		return err
	}

	if enable_www {

		key, err := StringVar(fs, "www-api-key")

		if err != nil {
			return err
		}

		if key == "xxxxxx" {

			warning := "-enable-www flag is set but -www-api-key is empty"

			if strict {
				return errors.New(warning)
			}

			log.Printf("[WARNING] %s\n", warning)
		}
	}

	return nil
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

	fs.String("spatialite-dsn", "", "A valid SQLite DSN for the '-cache spatialite/sqlite' or '-index spatialite' option. As of this writing for the '-index' and '-cache' options share the same '-spatailite' DSN.")
	fs.String("fs-path", "", "The root directory to look for features if '-cache fs'.")

	fs.Bool("is-wof", true, "Input data is WOF-flavoured GeoJSON. (Pass a value of '0' or 'false' if you need to index non-WOF documents.")

	// this is invoked/used in app/indexer.go but for the life of me I can't
	// figure out how to make the code in flags/exclude.go implement the
	// correct inferface wah wah so that flag.Lookup("exclude").Value returns
	// something we can loop over... so instead we just strings.Split() on
	// flag.Lookup("exclude").String() which is dumb but works...
	// (20180301/thisisaaronland)

	var exclude Exclude
	fs.Var(&exclude, "exclude", "Exclude (WOF) records based on their existential flags. Valid options are: ceased, deprecated, not-current, superseded.")

	fs.Bool("verbose", false, "Be chatty.")
	fs.Bool("strict", false, "Be strict about flags (and fail if any deprecated flags are used).")

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
