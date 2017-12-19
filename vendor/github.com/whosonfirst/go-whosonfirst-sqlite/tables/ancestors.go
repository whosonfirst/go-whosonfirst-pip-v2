package tables

import (
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	"github.com/whosonfirst/go-whosonfirst-sqlite/utils"
	"strings"
)

type AncestorsTable struct {
	sqlite.Table
	name string
}

type AncestorsRow struct {
	Id                int64
	AncestorID        int64
	AncestorPlacetype string
	LastModified      int64
}

func NewAncestorsTableWithDatabase(db sqlite.Database) (sqlite.Table, error) {

	t, err := NewAncestorsTable()

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func NewAncestorsTable() (sqlite.Table, error) {

	t := AncestorsTable{
		name: "ancestors",
	}

	return &t, nil
}

func (t *AncestorsTable) Name() string {
	return t.name
}

func (t *AncestorsTable) Schema() string {

	sql := `CREATE TABLE %s (
		id INTEGER NOT NULL,
		ancestor_id INTEGER NOT NULL,
		ancestor_placetype TEXT,
		lastmodified INTEGER
	);

	CREATE INDEX ancestors_by_id ON %s (id,ancestor_placetype,lastmodified);
	CREATE INDEX ancestors_by_ancestor ON %s (ancestor_id,ancestor_placetype,lastmodified);
	CREATE INDEX ancestors_by_lastmod ON %s (lastmodified);`

	return fmt.Sprintf(sql, t.Name(), t.Name(), t.Name(), t.Name())
}

func (t *AncestorsTable) InitializeTable(db sqlite.Database) error {

	return utils.CreateTableIfNecessary(db, t)
}

func (t *AncestorsTable) IndexFeature(db sqlite.Database, f geojson.Feature) error {

	conn, err := db.Conn()

	if err != nil {
		return err
	}

	tx, err := conn.Begin()

	id := f.Id()

	sql := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, t.Name())

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(id)

	if err != nil {
		return err
	}

	str_id := f.Id()

	hierarchies := whosonfirst.Hierarchies(f)
	lastmod := whosonfirst.LastModified(f)

	for _, h := range hierarchies {

		for pt_key, ancestor_id := range h {

			ancestor_placetype := strings.Replace(pt_key, "_id", "", -1)

			sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
				id, ancestor_id, ancestor_placetype, lastmodified
			) VALUES (
			  	 ?, ?, ?, ?
			)`, t.Name())

			stmt, err := tx.Prepare(sql)

			if err != nil {
				return err
			}

			defer stmt.Close()

			_, err = stmt.Exec(str_id, ancestor_id, ancestor_placetype, lastmod)

			if err != nil {
				return err
			}

		}

	}

	return tx.Commit()
}
