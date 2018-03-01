package app

import (
	"errors"
	"flag"
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/flags"
	"github.com/whosonfirst/go-whosonfirst-pip/index"
)

func NewApplicationIndex(fl *flag.FlagSet, appcache cache.Cache) (index.Index, error) {

	pip_index, err := flags.StringVar(fl, "pip-index")

	if err != nil {
		return nil, err
	}

	switch pip_index {
	case "rtree":
		return index.NewRTreeIndex(appcache)
	case "spatialite":

		db, err := NewSpatialiteDB(fl)

		if err != nil {
			return nil, err
		}

		return index.NewSpatialiteIndex(db, appcache)
	default:
		return nil, errors.New("Invalid engine")
	}
}
