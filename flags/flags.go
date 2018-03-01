package flags

import (
	"errors"
	"flag"
	_ "fmt"
	"os"
	"runtime"
)

func Parse(fl *flag.FlagSet, args []string) {

     	if len(args) > 0 && args[0] == "-h" {
		fl.Usage()
		os.Exit(0)
	}

	fl.Parse(args)
}

func Lookup(fl *flag.FlagSet, k string) (interface{}, error) {

	v := fl.Lookup(k)

	if v != nil {
		// Go is weird...
		return v.Value.(flag.Getter).Get(), nil
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

func NewFlagSet(name string) *flag.FlagSet {

	fs := flag.NewFlagSet(name, flag.ContinueOnError)

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
	fs.Int("processes", runtime.NumCPU()*2, "...")

	fs.String("spatialite-dsn", ":memory:", "...")
	fs.String("fs-path", "", "...")

	fs.Bool("is-wof", true, "...")

	// EXCLUDE FLAGS

	fs.Bool("verbose", false, "")

	return fs, nil
}
