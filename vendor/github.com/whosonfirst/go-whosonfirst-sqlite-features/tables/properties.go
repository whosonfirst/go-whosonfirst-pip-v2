package tables

import (
	"context"
	"fmt"
	"github.com/aaronland/go-sqlite"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features"
)

type PropertiesTableOptions struct {
	IndexAltFiles bool
}

func DefaultPropertiesTableOptions() (*PropertiesTableOptions, error) {

	opts := PropertiesTableOptions{
		IndexAltFiles: false,
	}

	return &opts, nil
}

type PropertiesTable struct {
	features.FeatureTable
	name    string
	options *PropertiesTableOptions
}

type PropertiesRow struct {
	Id           int64
	Body         string
	LastModified int64
}

func NewPropertiesTableWithDatabase(ctx context.Context, db sqlite.Database) (sqlite.Table, error) {

	opts, err := DefaultPropertiesTableOptions()

	if err != nil {
		return nil, err
	}

	return NewPropertiesTableWithDatabaseAndOptions(ctx, db, opts)
}

func NewPropertiesTableWithDatabaseAndOptions(ctx context.Context, db sqlite.Database, opts *PropertiesTableOptions) (sqlite.Table, error) {

	t, err := NewPropertiesTableWithOptions(ctx, opts)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func NewPropertiesTable(ctx context.Context) (sqlite.Table, error) {

	opts, err := DefaultPropertiesTableOptions()

	if err != nil {
		return nil, err
	}

	return NewPropertiesTableWithOptions(ctx, opts)
}

func NewPropertiesTableWithOptions(ctx context.Context, opts *PropertiesTableOptions) (sqlite.Table, error) {

	t := PropertiesTable{
		name:    "properties",
		options: opts,
	}

	return &t, nil
}

func (t *PropertiesTable) Name() string {
	return t.name
}

func (t *PropertiesTable) Schema() string {

	sql := `CREATE TABLE %s (
		id INTEGER NOT NULL,
		body TEXT,
		is_alt BOOLEAN,
		alt_label TEXT,
		lastmodified INTEGER
	);

	CREATE UNIQUE INDEX properties_by_id ON %s (id, alt_label);
	CREATE INDEX properties_by_alt ON %s (id, is_alt, alt_label);
	CREATE INDEX properties_by_lastmod ON %s (lastmodified);
	`

	return fmt.Sprintf(sql, t.Name(), t.Name(), t.Name(), t.Name())
}

func (t *PropertiesTable) InitializeTable(ctx context.Context, db sqlite.Database) error {

	return sqlite.CreateTableIfNecessary(ctx, db, t)
}

func (t *PropertiesTable) IndexRecord(ctx context.Context, db sqlite.Database, i interface{}) error {
	return t.IndexFeature(ctx, db, i.(geojson.Feature))
}

func (t *PropertiesTable) IndexFeature(ctx context.Context, db sqlite.Database, f geojson.Feature) error {

	conn, err := db.Conn()

	if err != nil {
		return err
	}

	str_id := f.Id()

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
		id, body, is_alt, alt_label, lastmodified
	) VALUES (
		?, ?, ?, ?, ?
	)`, t.Name())

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return err
	}

	defer stmt.Close()

	rsp_props := gjson.GetBytes(f.Bytes(), "properties")
	str_props := rsp_props.String()

	_, err = stmt.Exec(str_id, str_props, is_alt, alt_label, lastmod)

	if err != nil {
		return err
	}

	return tx.Commit()
}
