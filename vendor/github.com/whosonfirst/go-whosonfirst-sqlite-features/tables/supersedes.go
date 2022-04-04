package tables

import (
	"context"
	"fmt"
	"github.com/aaronland/go-sqlite"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features"
)

type SupersedesTable struct {
	features.FeatureTable
	name string
}

func NewSupersedesTableWithDatabase(ctx context.Context, db sqlite.Database) (sqlite.Table, error) {

	t, err := NewSupersedesTable(ctx)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func NewSupersedesTable(ctx context.Context) (sqlite.Table, error) {

	t := SupersedesTable{
		name: "supersedes",
	}

	return &t, nil
}

func (t *SupersedesTable) Name() string {
	return t.name
}

func (t *SupersedesTable) Schema() string {

	sql := `CREATE TABLE %s (
		id INTEGER NOT NULL,
		superseded_id INTEGER NOT NULL,
		superseded_by_id INTEGER NOT NULL,
		lastmodified INTEGER
	);

	CREATE UNIQUE INDEX supersedes_by ON %s (id,superseded_id, superseded_by_id);
	`

	return fmt.Sprintf(sql, t.Name(), t.Name())
}

func (t *SupersedesTable) InitializeTable(ctx context.Context, db sqlite.Database) error {

	return sqlite.CreateTableIfNecessary(ctx, db, t)
}

func (t *SupersedesTable) IndexRecord(ctx context.Context, db sqlite.Database, i interface{}) error {
	return t.IndexFeature(ctx, db, i.(geojson.Feature))
}

func (t *SupersedesTable) IndexFeature(ctx context.Context, db sqlite.Database, f geojson.Feature) error {

	is_alt := whosonfirst.IsAlt(f)

	if is_alt {
		return nil
	}

	conn, err := db.Conn()

	if err != nil {
		return err
	}

	tx, err := conn.Begin()

	if err != nil {
		return err
	}

	id := whosonfirst.Id(f)
	lastmod := whosonfirst.LastModified(f)

	sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
				id, superseded_id, superseded_by_id, lastmodified
			) VALUES (
			  	 ?, ?, ?, ?
			)`, t.Name())

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return err
	}

	defer stmt.Close()

	superseded_by := whosonfirst.SupersededBy(f)

	for _, other_id := range superseded_by {

		_, err = stmt.Exec(id, id, other_id, lastmod)

		if err != nil {
			return err
		}

	}

	supersedes := whosonfirst.Supersedes(f)

	for _, other_id := range supersedes {

		_, err = stmt.Exec(id, other_id, id, lastmod)

		if err != nil {
			return err
		}

	}

	return tx.Commit()
}
