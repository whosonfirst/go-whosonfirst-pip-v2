package app

import (
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/index"
)

type ApplicationIndexOptions struct {
     IncludeDeprecated bool
     IncludeSuperseded bool
     IncludeCeased bool
     IncludeNotCurrent bool
}

func DefaultApplicationIndexOptions() (ApplicationIndexOptions, error) {

	opts := ApplicationIndexOptions{
	     IncludeDeprecated: true,
	     IncludeSuperseded: true,
	     IncludeCeased: true,
	     IncludeNotCurrent: true,
	}

	return opts, nil
}

func ApplicationIndex(c cache.Cache) (index.Index, error) {

	return index.NewRTreeIndex(c)
}
