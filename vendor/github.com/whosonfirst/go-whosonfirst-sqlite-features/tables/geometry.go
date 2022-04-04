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

type GeometryTableOptions struct {
	IndexAltFiles bool
}

func DefaultGeometryTableOptions() (*GeometryTableOptions, error) {

	opts := GeometryTableOptions{
		IndexAltFiles: false,
	}

	return &opts, nil
}

type GeometryTable struct {
	features.FeatureTable
	name    string
	options *GeometryTableOptions
}

type GeometryRow struct {
	Id           int64
	Body         string
	LastModified int64
}

func NewGeometryTableWithDatabase(ctx context.Context, db sqlite.Database) (sqlite.Table, error) {

	opts, err := DefaultGeometryTableOptions()

	if err != nil {
		return nil, err
	}

	return NewGeometryTableWithDatabaseAndOptions(ctx, db, opts)
}

func NewGeometryTableWithDatabaseAndOptions(ctx context.Context, db sqlite.Database, opts *GeometryTableOptions) (sqlite.Table, error) {

	t, err := NewGeometryTableWithOptions(ctx, opts)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func NewGeometryTable(ctx context.Context) (sqlite.Table, error) {

	opts, err := DefaultGeometryTableOptions()

	if err != nil {
		return nil, err
	}

	return NewGeometryTableWithOptions(ctx, opts)
}

func NewGeometryTableWithOptions(ctx context.Context, opts *GeometryTableOptions) (sqlite.Table, error) {

	t := GeometryTable{
		name:    "geometry",
		options: opts,
	}

	return &t, nil
}

func (t *GeometryTable) Name() string {
	return t.name
}

func (t *GeometryTable) Schema() string {

	sql := `CREATE TABLE %s (
		id INTEGER NOT NULL,
		body TEXT,
		is_alt BOOLEAN,
		alt_label TEXT,
		lastmodified INTEGER
	);

	CREATE UNIQUE INDEX geometry_by_id ON %s (id, alt_label);
	CREATE INDEX geometry_by_alt ON %s (id, is_alt, alt_label);
	CREATE INDEX geometry_by_lastmod ON %s (lastmodified);
	`

	return fmt.Sprintf(sql, t.Name(), t.Name(), t.Name(), t.Name())
}

func (t *GeometryTable) InitializeTable(ctx context.Context, db sqlite.Database) error {

	return sqlite.CreateTableIfNecessary(ctx, db, t)
}

func (t *GeometryTable) IndexRecord(ctx context.Context, db sqlite.Database, i interface{}) error {
	return t.IndexFeature(ctx, db, i.(geojson.Feature))
}

func (t *GeometryTable) IndexFeature(ctx context.Context, db sqlite.Database, f geojson.Feature) error {

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

	rsp_geom := gjson.GetBytes(f.Bytes(), "geometry")
	str_geom := rsp_geom.String()

	_, err = stmt.Exec(str_id, str_geom, is_alt, alt_label, lastmod)

	if err != nil {
		return err
	}

	return tx.Commit()
}
