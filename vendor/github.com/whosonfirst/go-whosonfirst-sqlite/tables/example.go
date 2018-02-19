package tables

import (
	"encoding/json"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	"github.com/whosonfirst/go-whosonfirst-sqlite/utils"
)

type ExampleTable struct {
	sqlite.Table
	name string
}

func NewExampleTableWithDatabase(db sqlite.Database) (sqlite.Table, error) {

	t, err := NewExampleTable()

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func NewExampleTable() (sqlite.Table, error) {

	t := ExampleTable{
		name: "example",
	}

	return &t, nil
}

func (t *ExampleTable) Name() string {
	return t.name
}

func (t *ExampleTable) Schema() string {

	sql := `CREATE TABLE %s (
		id INTEGER NOT NULL PRIMARY KEY,
		body TEXT
	);`

	return fmt.Sprintf(sql, t.Name(), t.Name())
}

func (t *ExampleTable) InitializeTable(db sqlite.Database) error {

	return utils.CreateTableIfNecessary(db, t)
}

func (t *ExampleTable) IndexRecord(db sqlite.Database, i interface{}) error {

	conn, err := db.Conn()

	if err != nil {
		return err
	}

	tx, err := conn.Begin()

	if err != nil {
		return err
	}

	b, err := json.Marshal(i)

	if err != nil {
		return err
	}

	id := "FIX ME"
	body := string(b)

	sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
		id, body
	) VALUES (
		?, ?
	)`, t.Name())

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(id, body)

	if err != nil {
		return err
	}

	return tx.Commit()
}
