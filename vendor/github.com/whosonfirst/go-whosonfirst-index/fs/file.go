package fs

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-index"
)

func init() {
	dr := NewFileDriver()
	index.Register("file", dr)
}

func NewFileDriver() index.Driver {
	return &FileDriver{}
}

type FileDriver struct {
	index.Driver
}

func (d *FileDriver) Open(uri string) error {
	return nil
}

func (d *FileDriver) IndexURI(ctx context.Context, index_cb index.IndexerFunc, uri string) error {

	fh, err := readerFromPath(uri)

	if err != nil {
		return err
	}

	defer fh.Close()

	ctx = index.AssignPathContext(ctx, uri)
	return index_cb(ctx, fh)

}
