package tables

import (
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features"
	"github.com/whosonfirst/go-whosonfirst-sqlite/utils"
)

type GeoJSONTableOptions struct {
	IndexAltFiles bool
}

func DefaultGeoJSONTableOptions() (*GeoJSONTableOptions, error) {

	opts := GeoJSONTableOptions{
		IndexAltFiles: false,
	}

	return &opts, nil
}

type GeoJSONTable struct {
	features.FeatureTable
	name    string
	options *GeoJSONTableOptions
}

type GeoJSONRow struct {
	Id           int64
	Body         string
	LastModified int64
}

func NewGeoJSONTableWithDatabase(db sqlite.Database) (sqlite.Table, error) {

	opts, err := DefaultGeoJSONTableOptions()

	if err != nil {
		return nil, err
	}

	return NewGeoJSONTableWithDatabaseAndOptions(db, opts)
}

func NewGeoJSONTableWithDatabaseAndOptions(db sqlite.Database, opts *GeoJSONTableOptions) (sqlite.Table, error) {

	t, err := NewGeoJSONTableWithOptions(opts)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func NewGeoJSONTable() (sqlite.Table, error) {

	opts, err := DefaultGeoJSONTableOptions()

	if err != nil {
		return nil, err
	}

	return NewGeoJSONTableWithOptions(opts)
}

func NewGeoJSONTableWithOptions(opts *GeoJSONTableOptions) (sqlite.Table, error) {

	t := GeoJSONTable{
		name:    "geojson",
		options: opts,
	}

	return &t, nil
}

func (t *GeoJSONTable) Name() string {
	return t.name
}

func (t *GeoJSONTable) Schema() string {

	sql := `CREATE TABLE %s (
		id INTEGER NOT NULL,
		body TEXT,
		source TEXT,
		is_alt BOOLEAN,
		lastmodified INTEGER
	);

	CREATE UNIQUE INDEX geojson_by_id ON %s (id, source);
	CREATE INDEX geojson_by_alt ON %s (id, is_alt);
	CREATE INDEX geojson_by_lastmod ON %s (lastmodified);
	`

	return fmt.Sprintf(sql, t.Name(), t.Name(), t.Name(), t.Name())
}

func (t *GeoJSONTable) InitializeTable(db sqlite.Database) error {

	return utils.CreateTableIfNecessary(db, t)
}

func (t *GeoJSONTable) IndexRecord(db sqlite.Database, i interface{}) error {
	return t.IndexFeature(db, i.(geojson.Feature))
}

func (t *GeoJSONTable) IndexFeature(db sqlite.Database, f geojson.Feature) error {

	conn, err := db.Conn()

	if err != nil {
		return err
	}

	str_id := f.Id()
	body := f.Bytes()

	source := whosonfirst.Source(f)
	is_alt := whosonfirst.IsAlt(f)

	if is_alt && !t.options.IndexAltFiles {
		return nil
	}

	lastmod := whosonfirst.LastModified(f)

	tx, err := conn.Begin()

	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
		id, body, source, is_alt, lastmodified
	) VALUES (
		?, ?, ?, ?, ?
	)`, t.Name())

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return err
	}

	defer stmt.Close()

	str_body := string(body)

	_, err = stmt.Exec(str_id, str_body, source, is_alt, lastmod)

	if err != nil {
		return err
	}

	return tx.Commit()
}
