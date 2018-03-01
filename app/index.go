package app

import (
	"errors"
	"flag"
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/flags"
	"github.com/whosonfirst/go-whosonfirst-pip/index"
)

func NewApplicationIndex(fl *fl.FlagSet, appcache cache.Cache) (index.Index, error) {

	pip_index := flags.Lookup(fl, "pip-index")

	if err != nil {
		return nil, err
	}

	switch pip_index {
	case "rtree":
		return index.NewRTreeIndex(appcache)
	case "spatialite":
		return index.NewSpatialiteIndex(db, appcache)
	default:
		return nil, errors.New("Invalid engine")
	}
}
