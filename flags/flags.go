package flags

import (
	"errors"
	"flag"
	"fmt"
	_ "log"
	"os"
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

	fs.String("index", "rtree", "Valid options are: rtree, spatialite")
	fs.String("cache", "gocache", "Valid options are: gocache, fs, spatialite")

	fs.String("mode", "files", "...")

	fs.String("spatialite-dsn", ":memory:", "...")
	fs.String("fs-path", "", "...")

	fs.Bool("is-wof", true, "Input data is WOF-flavoured GeoJSON")

	fs.Bool("enable-extras", false, "")
	fs.String("extras-dsn", ":tmpfile:", "")

	// this is invoked/used in app/indexer.go but for the life of me I can't
	// figure out how to make the code in flags/exclude.go implement the
	// correct inferface wah wah so that flag.Lookup("exclude").Value returns
	// something we can loop over... so instead we just strings.Split() on
	// flag.Lookup("exclude").String() which is dumb but works...
	// (20180301/thisisaaronland)

	var exclude Exclude
	fs.Var(&exclude, "exclude", "Exclude (WOF) records based on their existential flags. Valid options are: ceased, deprecated, not-current, superseded.")

	fs.Bool("verbose", false, "")

	return fs, nil
}
