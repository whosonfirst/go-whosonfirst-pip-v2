package app

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/geometry"
	"github.com/whosonfirst/go-whosonfirst-index"
	pip "github.com/whosonfirst/go-whosonfirst-pip/index"
	pip_utils "github.com/whosonfirst/go-whosonfirst-pip/utils"
	"io"
	_ "log"
)

type ApplicationIndexerOptions struct {
	IndexMode string
	IsWOF     bool
}

func DefaultApplicationIndexerOptions() (ApplicationIndexerOptions, error) {

	opts := ApplicationIndexerOptions{
		IndexMode: "",
		IsWOF:     true,
	}

	return opts, nil
}

func NewApplicationIndexer(appindex pip.Index, opts ApplicationIndexerOptions) (*index.Indexer, error) {

	cb := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

		var f geojson.Feature

		if opts.IsWOF {

			ok, err := pip_utils.IsValidRecord(fh, ctx)

			if err != nil {
				return err
			}

			if !ok {
				return err
			}

			tmp, err := feature.LoadWOFFeatureFromReader(fh)

			if err != nil {
				return err
			}

			f = tmp

		} else {

			tmp, err := feature.LoadFeatureFromReader(fh)

			if err != nil {
				return err
			}

			f = tmp
		}

		geom_type := geometry.Type(f)

		if geom_type == "Point" {
			return nil
		}

		return appindex.IndexFeature(f)
	}

	return index.NewIndexer(opts.IndexMode, cb)
}
