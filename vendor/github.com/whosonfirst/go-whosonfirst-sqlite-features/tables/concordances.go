package tables

import (
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features"
	"github.com/whosonfirst/go-whosonfirst-sqlite/utils"
)

type ConcordancesTable struct {
	features.FeatureTable
	name string
}

type ConcordancesRow struct {
	Id           int64
	OtherID      string
	OtherSource  string
	LastModified int64
}

func NewConcordancesTableWithDatabase(db sqlite.Database) (sqlite.Table, error) {

	t, err := NewConcordancesTable()

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func NewConcordancesTable() (sqlite.Table, error) {

	t := ConcordancesTable{
		name: "concordances",
	}

	return &t, nil
}

func (t *ConcordancesTable) Name() string {
	return t.name
}

func (t *ConcordancesTable) Schema() string {

	sql := `CREATE TABLE %s (
		id INTEGER NOT NULL,
		other_id INTEGER NOT NULL,
		other_source TEXT,
		lastmodified INTEGER
	);

	CREATE INDEX concordances_by_id ON %s (id,lastmodified);
	CREATE INDEX concordances_by_other_id ON %s (other_source,other_id);	
	CREATE INDEX concordances_by_other_lastmod ON %s (other_source,other_id,lastmodified);
	CREATE INDEX concordances_by_lastmod ON %s (lastmodified);`

	return fmt.Sprintf(sql, t.Name(), t.Name(), t.Name(), t.Name(), t.Name())
}

func (t *ConcordancesTable) InitializeTable(db sqlite.Database) error {

	return utils.CreateTableIfNecessary(db, t)
}

func (t *ConcordancesTable) IndexRecord(db sqlite.Database, i interface{}) error {
	return t.IndexFeature(db, i.(geojson.Feature))
}

func (t *ConcordancesTable) IndexFeature(db sqlite.Database, f geojson.Feature) error {

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

	concordances, err := whosonfirst.Concordances(f)

	if err != nil {
		return err
	}

	lastmod := whosonfirst.LastModified(f)

	for other_source, other_id := range concordances {

		sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
				id, other_id, other_source, lastmodified
			) VALUES (
			  	 ?, ?, ?, ?
			)`, t.Name())

		stmt, err := tx.Prepare(sql)

		if err != nil {
			return err
		}

		defer stmt.Close()

		_, err = stmt.Exec(str_id, other_id, other_source, lastmod)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
