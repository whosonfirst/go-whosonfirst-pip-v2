package flags

import (
	"flag"
	"runtime"
)

func Lookup(fl *flag.FlagSet, key string) (interface{}, error) {

     v := fl.Lookup(k)

     if v != nil {
     	return v.Value, nil
     }

     return nil, errors.New("Unknown flag")
}

func CommonFlags() (*flag.FlagSet, error) {

	common := flag.NewFlagSet("common", flag.PanicOnError)

	common.String("index", "rtree", "Valid options are: rtree, spatialite")
	common.String("cache", "gocache", "Valid options are: gocache, fs, spatialite")

	common.String("mode", "files", "...")
	common.Int("processes", runtime.NumCPU()*2, "...")

	common.String("spatialite-dsn", ":memory:", "...")
	common.String("fs-path", "", "...")

	fl.Bool("is-wof", true, "...")

	// EXCLUDE FLAGS

	common.Bool("verbose", false, "")

	// MAYBE ?
	common.Int("polylines-coords", 100, "...")

	return common, nil
}
