package flags

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
)

func Parse(fl *flag.FlagSet, args []string) {

	fl.Parse(args)

	help, _ := BoolVar(fl, "h")

	if !help {
		help, _ = BoolVar(fl, "help")
	}

	if help {
		fl.Usage()
		os.Exit(0)
	}
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

	fs.Bool("h", false, "")
	fs.Bool("help", false, "")

	fs.Usage = func() {
		fmt.Println("GO IS WEIRD, PART 2")
	}

	return fs
}

func CommonFlags() (*flag.FlagSet, error) {

	common := NewFlagSet("common")

	common.String("index", "rtree", "Valid options are: rtree, spatialite")
	common.String("cache", "gocache", "Valid options are: gocache, fs, spatialite")

	common.String("mode", "files", "...")
	common.Int("processes", runtime.NumCPU()*2, "...")

	common.String("spatialite-dsn", ":memory:", "...")
	common.String("fs-path", "", "...")

	common.Bool("is-wof", true, "...")

	// EXCLUDE FLAGS

	common.Bool("verbose", false, "")

	return common, nil
}
