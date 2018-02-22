package utils

import (
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	_ "log"
	"os"
)

func HasTable(db sqlite.Database, table string) (bool, error) {

	// PLEASE MEMOIZE ME... (20180221/thisisaaronland)

	check_tables := true
	has_table := false

	dsn := db.DSN()

	if dsn != ":memory:" {

		_, err := os.Stat(dsn)

		if os.IsNotExist(err) {
			check_tables = false
			has_table = false
		}
	}

	if check_tables {

		conn, err := db.Conn()

		if err != nil {
			return false, err
		}

		sql := "SELECT name FROM sqlite_master WHERE type='table'"

		rows, err := conn.Query(sql)

		if err != nil {
			return false, err
		}

		defer rows.Close()

		for rows.Next() {

			var name string
			err := rows.Scan(&name)

			if err != nil {
				return false, err
			}

			if name == table {
				has_table = true
				break
			}
		}
	}

	return has_table, nil
}

func CreateTableIfNecessary(db sqlite.Database, t sqlite.Table) error {

	create := false

	has_table, err := HasTable(db, t.Name())

	if err != nil {
		return err
	}

	if !has_table {
		create = true
	}

	if create {

		sql := t.Schema()

		conn, err := db.Conn()

		if err != nil {
			return err
		}

		_, err = conn.Exec(sql)

		if err != nil {
			return err
		}

	}

	return nil
}
