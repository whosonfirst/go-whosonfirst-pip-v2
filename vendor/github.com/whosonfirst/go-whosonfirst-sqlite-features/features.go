package features

import (
       "github.com/whosonfirst/go-whosonfirst-sqlite"
       "github.com/whosonfirst/go-whosonfirst-geojson-v2"
)

type FeatureTable interface {
     sqlite.Table
     IndexFeature(sqlite.Database, geojson.Feature) error
}
