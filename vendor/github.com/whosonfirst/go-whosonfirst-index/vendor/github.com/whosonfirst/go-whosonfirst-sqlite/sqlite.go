package sqlite

import (
       "database/sql"
       "github.com/whosonfirst/go-whosonfirst-geojson-v2"
)

type Database interface {
     Conn() (*sql.DB, error)
     DSN() string
     Close() error
}

type Table interface {
     Name() string
     Schema() string
     InitializeTable(Database) error
     IndexFeature(Database, geojson.Feature) error
}

// this is here so we can pass both sql.Row and sql.Rows to the
// ResultSetFunc below (20170824/thisisaaronland)

type ResultSet interface {
	Scan(dest ...interface{}) error
}

type ResultRow interface {
     Row() interface{}
}

type ResultSetFunc func(row ResultSet) (ResultRow, error)
