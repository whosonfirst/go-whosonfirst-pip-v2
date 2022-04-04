package tables

import (
	"context"
	"fmt"
	"github.com/aaronland/go-sqlite"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features"
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

func NewGeoJSONTableWithDatabase(ctx context.Context, db sqlite.Database) (sqlite.Table, error) {

	opts, err := DefaultGeoJSONTableOptions()

	if err != nil {
		return nil, err
	}

	return NewGeoJSONTableWithDatabaseAndOptions(ctx, db, opts)
}

func NewGeoJSONTableWithDatabaseAndOptions(ctx context.Context, db sqlite.Database, opts *GeoJSONTableOptions) (sqlite.Table, error) {

	t, err := NewGeoJSONTableWithOptions(ctx, opts)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func NewGeoJSONTable(ctx context.Context) (sqlite.Table, error) {

	opts, err := DefaultGeoJSONTableOptions()

	if err != nil {
		return nil, err
	}

	return NewGeoJSONTableWithOptions(ctx, opts)
}

func NewGeoJSONTableWithOptions(ctx context.Context, opts *GeoJSONTableOptions) (sqlite.Table, error) {

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
		alt_label TEXT,
		lastmodified INTEGER
	);

	CREATE UNIQUE INDEX geojson_by_id ON %s (id, source, alt_label);
	CREATE INDEX geojson_by_alt ON %s (id, is_alt, alt_label);
	CREATE INDEX geojson_by_lastmod ON %s (lastmodified);
	`

	return fmt.Sprintf(sql, t.Name(), t.Name(), t.Name(), t.Name())
}

func (t *GeoJSONTable) InitializeTable(ctx context.Context, db sqlite.Database) error {

	return sqlite.CreateTableIfNecessary(ctx, db, t)
}

func (t *GeoJSONTable) IndexRecord(ctx context.Context, db sqlite.Database, i interface{}) error {
	return t.IndexFeature(ctx, db, i.(geojson.Feature))
}

func (t *GeoJSONTable) IndexFeature(ctx context.Context, db sqlite.Database, f geojson.Feature) error {

	conn, err := db.Conn()

	if err != nil {
		return err
	}

	str_id := f.Id()
	body := f.Bytes()

	source := whosonfirst.Source(f)
	is_alt := whosonfirst.IsAlt(f)
	alt_label := whosonfirst.AltLabel(f)

	if is_alt && !t.options.IndexAltFiles {
		return nil
	}

	lastmod := whosonfirst.LastModified(f)

	tx, err := conn.Begin()

	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
		id, body, source, is_alt, alt_label, lastmodified
	) VALUES (
		?, ?, ?, ?, ?, ?
	)`, t.Name())

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return err
	}

	defer stmt.Close()

	str_body := string(body)

	_, err = stmt.Exec(str_id, str_body, source, is_alt, alt_label, lastmod)

	if err != nil {
		return err
	}

	return tx.Commit()
}
