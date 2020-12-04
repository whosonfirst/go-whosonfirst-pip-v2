package fs

import (
	"bufio"
	"context"
	"github.com/whosonfirst/go-whosonfirst-index"
)

func init() {
	dr := NewFileListDriver()
	index.Register("filelist", dr)
}

func NewFileListDriver() index.Driver {
	return &FileListDriver{}
}

type FileListDriver struct {
	index.Driver
}

func (d *FileListDriver) Open(uri string) error {
	return nil
}

func (d *FileListDriver) IndexURI(ctx context.Context, index_cb index.IndexerFunc, uri string) error {

	fh, err := readerFromPath(uri)

	if err != nil {
		return err
	}

	defer fh.Close()

	scanner := bufio.NewScanner(fh)

	for scanner.Scan() {

		select {
		case <-ctx.Done():
			break
		default:
			// pass
		}

		path := scanner.Text()

		fh, err := readerFromPath(path)

		if err != nil {
			return err
		}

		ctx = index.AssignPathContext(ctx, path)

		err = index_cb(ctx, fh)

		if err != nil {
			return err
		}
	}

	err = scanner.Err()

	if err != nil {
		return err
	}

	return nil
}
