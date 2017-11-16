package app

import (
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/index"
)

type ApplicationIndexOptions struct {
}

func DefaultApplicationIndexOptions() (ApplicationIndexOptions, error) {

	opts := ApplicationIndexOptions{}

	return opts, nil
}

func ApplicationIndex(c cache.Cache) (index.Index, error) {

	return index.NewRTreeIndex(c)
}
