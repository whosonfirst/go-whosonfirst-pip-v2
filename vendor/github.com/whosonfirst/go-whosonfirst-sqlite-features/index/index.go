package index

import (
	"context"
	"errors"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	wof_index "github.com/whosonfirst/go-whosonfirst-index"
 	wof_utils "github.com/whosonfirst/go-whosonfirst-index/utils"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	sql_index "github.com/whosonfirst/go-whosonfirst-sqlite/index"
	"github.com/whosonfirst/warning"
	"io"
)

// THIS IS A TOTAL HACK UNTIL WE CAN SORT THINGS OUT IN
// go-whosonfirst-index... (20180206/thisisaaronland)

type Closer struct {
	fh io.Reader
}

func (c Closer) Read(b []byte) (int, error) {
	return c.fh.Read(b)
}

func (c Closer) Close() error {
	return nil
}

func NewDefaultSQLiteFeaturesIndexer(db sqlite.Database, to_index []sqlite.Table) (*sql_index.SQLiteIndexer, error) {

	cb := func(ctx context.Context, fh io.Reader, args ...interface{}) (interface{}, error) {

		select {

		case <-ctx.Done():
			return nil, nil
		default:
			path, err := wof_index.PathForContext(ctx)

			if err != nil {
				return nil, err
			}

			ok, err := wof_utils.IsPrincipalWOFRecord(fh, ctx)

			if err != nil {
				return nil, err
			}

			if !ok {
				return nil, nil
			}

			// HACK - see above
			closer := Closer{fh}

			i, err := feature.LoadWOFFeatureFromReader(closer)

			if err != nil && !warning.IsWarning(err){
				msg := fmt.Sprintf("Unable to load %s, because %s", path, err)
				return nil, errors.New(msg)
			}

			return i, nil
		}
	}

	return sql_index.NewSQLiteIndexer(db, to_index, cb)
}
