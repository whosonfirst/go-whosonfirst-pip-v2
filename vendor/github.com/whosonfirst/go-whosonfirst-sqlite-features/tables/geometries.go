package tables

import (
	"fmt"
	"github.com/twpayne/go-geom"
	gogeom_geojson "github.com/twpayne/go-geom/encoding/geojson"
	"github.com/twpayne/go-geom/encoding/wkt"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/geometry"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features"
	"github.com/whosonfirst/go-whosonfirst-sqlite/utils"
	_ "log"
)

type GeometriesTableOptions struct {
	IndexAltFiles bool
}

func DefaultGeometriesTableOptions() (*GeometriesTableOptions, error) {

	opts := GeometriesTableOptions{
		IndexAltFiles: false,
	}

	return &opts, nil
}

type GeometriesTable struct {
	features.FeatureTable
	name    string
	options *GeometriesTableOptions
}

type GeometriesRow struct {
	Id           int64
	Body         string
	LastModified int64
}

func NewGeometriesTable() (sqlite.Table, error) {

	opts, err := DefaultGeometriesTableOptions()

	if err != nil {
		return nil, err
	}

	return NewGeometriesTableWithOptions(opts)
}

func NewGeometriesTableWithOptions(opts *GeometriesTableOptions) (sqlite.Table, error) {

	t := GeometriesTable{
		name:    "geometries",
		options: opts,
	}

	return &t, nil
}

func NewGeometriesTableWithDatabase(db sqlite.Database) (sqlite.Table, error) {

	opts, err := DefaultGeometriesTableOptions()

	if err != nil {
		return nil, err
	}

	return NewGeometriesTableWithDatabaseAndOptions(db, opts)
}

func NewGeometriesTableWithDatabaseAndOptions(db sqlite.Database, opts *GeometriesTableOptions) (sqlite.Table, error) {

	t, err := NewGeometriesTableWithOptions(opts)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func (t *GeometriesTable) Name() string {
	return t.name
}

func (t *GeometriesTable) Schema() string {

	// really this should probably be the SPR table + geom but
	// let's just get this working first and then make it fancy
	// (20180109/thisisaaronland)

	// https://www.gaia-gis.it/spatialite-1.0a/SpatiaLite-tutorial.html
	// http://www.gaia-gis.it/gaia-sins/spatialite-sql-4.3.0.html

	// Note the InitSpatialMetaData() command because this:
	// https://stackoverflow.com/questions/17761089/cannot-create-column-with-spatialite-unexpected-metadata-layout

	sql := `CREATE TABLE %s (
		id INTEGER NOT NULL PRIMARY KEY,
		is_alt TINYINT,
		type TEXT,
		lastmodified INTEGER
	);

	SELECT InitSpatialMetaData();
	SELECT AddGeometryColumn('%s', 'geom', 4326, 'GEOMETRY', 'XY');
	SELECT CreateSpatialIndex('%s', 'geom');

	CREATE INDEX geometries_by_lastmod ON %s (lastmodified);`

	return fmt.Sprintf(sql, t.Name(), t.Name(), t.Name(), t.Name())
}

func (t *GeometriesTable) InitializeTable(db sqlite.Database) error {

	return utils.CreateTableIfNecessary(db, t)
}

func (t *GeometriesTable) IndexRecord(db sqlite.Database, i interface{}) error {
	return t.IndexFeature(db, i.(geojson.Feature))
}

func (t *GeometriesTable) IndexFeature(db sqlite.Database, f geojson.Feature) error {

	conn, err := db.Conn()

	if err != nil {
		return err
	}

	str_id := f.Id()
	is_alt := whosonfirst.IsAlt(f)

	if is_alt && !t.options.IndexAltFiles {
		return nil
	}

	lastmod := whosonfirst.LastModified(f)

	tx, err := conn.Begin()

	if err != nil {
		return err
	}

	str_geom, err := geometry.ToString(f)

	if err != nil {
		return err
	}

	// but wait! there's more!! for reasons I've forgotten (simonw told me)
	// the spatialite doesn't really like indexing GeomFromGeoJSON but also
	// doesn't complain about it - it just chugs along happily filling your
	// database with null geometries so we're going to take advantage of the
	// handy "go-geom" package to convert the GeoJSON geometry in to WKT -
	// it is "one more thing" to import and maybe it would be better to just
	// write a custom converter but not today...
	// (20180122/thisisaaronland)

	var g geom.T
	err = gogeom_geojson.Unmarshal([]byte(str_geom), &g)

	if err != nil {
		return err
	}

	str_wkt, err := wkt.Marshal(g)

	sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
		id, is_alt, type, geom, lastmodified
	) VALUES (
		?, ?, ?, GeomFromText('%s', 4326), ?
	)`, t.Name(), str_wkt)

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return err
	}

	defer stmt.Close()

	geom_type := "common"

	_, err = stmt.Exec(str_id, is_alt, geom_type, lastmod)

	if err != nil {
		return err
	}

	return tx.Commit()
}
