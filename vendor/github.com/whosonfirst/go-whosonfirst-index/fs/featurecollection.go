package fs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-index"
	"io/ioutil"
)

func init() {
	dr := NewFeatureCollectionDriver()
	index.Register("featurecollection", dr)
}

func NewFeatureCollectionDriver() index.Driver {
	return &FeatureCollectionDriver{}
}

type FeatureCollectionDriver struct {
	index.Driver
}

func (d *FeatureCollectionDriver) Open(uri string) error {
	return nil
}

func (d *FeatureCollectionDriver) IndexURI(ctx context.Context, index_cb index.IndexerFunc, uri string) error {

	fh, err := readerFromPath(uri)

	if err != nil {
		return err
	}

	defer fh.Close()

	body, err := ioutil.ReadAll(fh)

	if err != nil {
		return err
	}

	type FC struct {
		Type     string
		Features []interface{}
	}

	var collection FC

	err = json.Unmarshal(body, &collection)

	if err != nil {
		return err
	}

	for i, f := range collection.Features {

		select {
		case <-ctx.Done():
			break
		default:
			// pass
		}

		feature, err := json.Marshal(f)

		if err != nil {
			return err
		}

		fh := bytes.NewBuffer(feature)

		path := fmt.Sprintf("%s#%d", uri, i)
		ctx = index.AssignPathContext(ctx, path)

		err = index_cb(ctx, fh)

		if err != nil {
			return err
		}
	}

	return nil
}
