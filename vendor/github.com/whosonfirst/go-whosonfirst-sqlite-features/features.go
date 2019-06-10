package features

import (
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
)

type FeatureTable interface {
	sqlite.Table
	IndexFeature(sqlite.Database, geojson.Feature) error
}
