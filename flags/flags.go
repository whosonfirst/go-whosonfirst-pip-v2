package flags

import (
	"flag"
	"runtime"
)

func CommonFlags() (*flag.FlagSet, error) {

	common := flag.NewFlagSet("common", flag.PanicOnError)

	common.String("index", "rtree", "Valid options are: rtree, spatialite")
	common.String("cache", "gocache", "Valid options are: gocache, fs, spatialite")

	common.String("mode", "files", "...")
	common.Int("processes", runtime.NumCPU()*2, "...")

	common.Bool("is-wof", true, "...")

	common.Bool("enable-geojson", false, "Allow users to request GeoJSON FeatureCollection formatted responses.")
	common.Bool("enable-extras", false, "")
	common.Bool("enable-candidates", false, "")
	common.Bool("enable-polylines", false, "")
	common.Bool("enable-www", false, "")

	common.Bool("verbose", false, "")

	return common, nil
}
