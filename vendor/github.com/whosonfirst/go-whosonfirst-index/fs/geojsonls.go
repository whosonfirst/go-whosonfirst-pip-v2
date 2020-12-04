package fs

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-index"
	"io"
)

func init() {
	dr := NewGeoJSONLDriver()
	index.Register("geojsonl", dr)
}

func NewGeoJSONLDriver() index.Driver {
	return &GeojsonLDriver{}
}

type GeojsonLDriver struct {
	index.Driver
}

func (d *GeojsonLDriver) Open(uri string) error {
	return nil
}

func (d *GeojsonLDriver) IndexURI(ctx context.Context, index_cb index.IndexerFunc, uri string) error {

	fh, err := readerFromPath(uri)

	if err != nil {
		return err
	}

	defer fh.Close()

	// see this - we're using ReadLine because it's entirely possible
	// that the raw GeoJSON (LS) will be too long for bufio.Scanner
	// see also - https://golang.org/pkg/bufio/#Reader.ReadLine
	// (20170822/thisisaaronland)

	reader := bufio.NewReader(fh)
	raw := bytes.NewBuffer([]byte(""))

	i := 0

	for {

		select {
		case <-ctx.Done():
			break
		default:
			// pass
		}

		path := fmt.Sprintf("%s#%d", uri, i)
		i += 1

		fragment, is_prefix, err := reader.ReadLine()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		raw.Write(fragment)

		if is_prefix {
			continue
		}

		fh := bytes.NewReader(raw.Bytes())

		ctx = index.AssignPathContext(ctx, path)
		err = index_cb(ctx, fh)

		if err != nil {
			return err
		}

		raw.Reset()
	}

	return nil
}
