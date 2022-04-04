package tables

// https://www.sqlite.org/rtree.html

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aaronland/go-sqlite"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/geometry"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features"
	_ "log"
)

type RTreeTableOptions struct {
	IndexAltFiles bool
}

func DefaultRTreeTableOptions() (*RTreeTableOptions, error) {

	opts := RTreeTableOptions{
		IndexAltFiles: false,
	}

	return &opts, nil
}

type RTreeTable struct {
	features.FeatureTable
	name    string
	options *RTreeTableOptions
}

func NewRTreeTable(ctx context.Context) (sqlite.Table, error) {

	opts, err := DefaultRTreeTableOptions()

	if err != nil {
		return nil, err
	}

	return NewRTreeTableWithOptions(ctx, opts)
}

func NewRTreeTableWithOptions(ctx context.Context, opts *RTreeTableOptions) (sqlite.Table, error) {

	t := RTreeTable{
		name:    "rtree",
		options: opts,
	}

	return &t, nil
}

func NewRTreeTableWithDatabase(ctx context.Context, db sqlite.Database) (sqlite.Table, error) {

	opts, err := DefaultRTreeTableOptions()

	if err != nil {
		return nil, err
	}

	return NewRTreeTableWithDatabaseAndOptions(ctx, db, opts)
}

func NewRTreeTableWithDatabaseAndOptions(ctx context.Context, db sqlite.Database, opts *RTreeTableOptions) (sqlite.Table, error) {

	t, err := NewRTreeTableWithOptions(ctx, opts)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func (t *RTreeTable) Name() string {
	return t.name
}

func (t *RTreeTable) Schema() string {

	/*

		3.1.1. Column naming details

		In the argments to "rtree" in the CREATE VIRTUAL TABLE statement, the names of the columns are taken from the first token of each argument. All subsequent tokens within each argument are silently ignored. This means, for example, that if you try to give a column a type affinity or add a constraint such as UNIQUE or NOT NULL or DEFAULT to a column, those extra tokens are accepted as valid, but they do not change the behavior of the rtree. In an RTREE virtual table, the first column always has a type affinity of INTEGER and all other data columns have a type affinity of NUMERIC.

		Recommended practice is to omit any extra tokens in the rtree specification. Let each argument to "rtree" be a single ordinary label that is the name of the corresponding column, and omit all other tokens from the argument list.

		4.1. Auxiliary Columns

		Beginning with SQLite version 3.24.0 (2018-06-04), r-tree tables can have auxiliary columns that store arbitrary data. Auxiliary columns can be used in place of secondary tables such as "demo_data".

		Auxiliary columns are marked with a "+" symbol before the column name. Auxiliary columns must come after all of the coordinate boundary columns. There is a limit of no more than 100 auxiliary columns. The following example shows an r-tree table with auxiliary columns that is equivalent to the two tables "demo_index" and "demo_data" above:

		Note: Auxiliary columns must come at the end of a table definition
	*/

	sql := `CREATE VIRTUAL TABLE %s USING rtree (
		id,
		min_x,
		max_x,
		min_y,
		max_y,
		+wof_id INTEGER,
		+is_alt TINYINT,
		+alt_label TEXT,
		+geometry BLOB,
		+lastmodified INTEGER
	);`

	return fmt.Sprintf(sql, t.Name())
}

func (t *RTreeTable) InitializeTable(ctx context.Context, db sqlite.Database) error {

	return sqlite.CreateTableIfNecessary(ctx, db, t)
}

func (t *RTreeTable) IndexRecord(ctx context.Context, db sqlite.Database, i interface{}) error {
	return t.IndexFeature(ctx, db, i.(geojson.Feature))
}

func (t *RTreeTable) IndexFeature(ctx context.Context, db sqlite.Database, f geojson.Feature) error {

	switch geometry.Type(f) {
	case "Polygon", "MultiPolygon":
		// pass
	default:
		return nil
	}

	conn, err := db.Conn()

	if err != nil {
		return err
	}

	wof_id := f.Id()
	is_alt := whosonfirst.IsAlt(f) // this returns a boolean which is interpreted as a float by SQLite

	if is_alt && !t.options.IndexAltFiles {
		return nil
	}

	alt_label := ""

	if is_alt {

		alt_label = whosonfirst.AltLabel(f)

		if alt_label == "" {
			return errors.New("Missing src:alt_label property")
		}
	}

	lastmod := whosonfirst.LastModified(f)

	polygons, err := f.Polygons()

	if err != nil {
		return err
	}

	tx, err := conn.Begin()

	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
		id, min_x, max_x, min_y, max_y, wof_id, is_alt, alt_label, geometry, lastmodified
	) VALUES (
		NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?
	)`, t.Name())

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return err
	}

	defer stmt.Close()

	// this should be updated to use go-whosonfirst-geojson-v2/geometry GeometryForFeature
	// so that we're not translating between [][][]float64 and skleterjohn/geom things
	// twice (20201214/thisisaaronland)

	for _, poly := range polygons {

		exterior_ring := poly.ExteriorRing()
		bbox := exterior_ring.Bounds()

		sw := bbox.Min
		ne := bbox.Max

		points := make([][][]float64, 0)

		exterior_points := make([][]float64, 0)

		for _, c := range exterior_ring.Vertices() {
			pt := []float64{c.X, c.Y}
			exterior_points = append(exterior_points, pt)
		}

		points = append(points, exterior_points)

		for _, interior_ring := range poly.InteriorRings() {

			interior_points := make([][]float64, 0)

			for _, c := range interior_ring.Vertices() {
				pt := []float64{c.X, c.Y}
				interior_points = append(interior_points, pt)
			}

			points = append(points, interior_points)
		}

		points_enc, err := json.Marshal(points)

		if err != nil {
			return err
		}

		_, err = stmt.Exec(sw.X, ne.X, sw.Y, ne.Y, wof_id, is_alt, alt_label, string(points_enc), lastmod)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
