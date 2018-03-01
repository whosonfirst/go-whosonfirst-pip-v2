package flags

import (
	"errors"
	"flag"
	"runtime"
)

func Lookup(fl *flag.FlagSet, k string) (interface{}, error) {

	v := fl.Lookup(k)

	if v != nil {
		return v.Value, nil
	}

	return nil, errors.New("Unknown flag")
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

func CommonFlags() (*flag.FlagSet, error) {

	common := flag.NewFlagSet("common", flag.PanicOnError)

	common.String("index", "rtree", "Valid options are: rtree, spatialite")
	common.String("cache", "gocache", "Valid options are: gocache, fs, spatialite")

	common.String("mode", "files", "...")
	common.Int("processes", runtime.NumCPU()*2, "...")

	common.String("spatialite-dsn", ":memory:", "...")
	common.String("fs-path", "", "...")

	common.Bool("is-wof", true, "...")

	// EXCLUDE FLAGS

	common.Bool("verbose", false, "")

	// MAYBE ?
	common.Int("polylines-coords", 100, "...")

	return common, nil
}
