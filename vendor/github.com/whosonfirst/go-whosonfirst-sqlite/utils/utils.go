package utils

import (
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	_ "log"
	"os"
	"regexp"
)

var re_mem *regexp.Regexp
var re_file *regexp.Regexp
var lookup_table map[string]bool

func init() {
	re_mem = regexp.MustCompile(`^(file\:)?\:memory\:.*`)
	re_file = regexp.MustCompile(`^file\:([^\?]+)(?:\?.*)?$`)

	lookup_table = make(map[string]bool)
}

func HasTable(db sqlite.Database, table string) (bool, error) {

	dsn := db.DSN()

	lookup_key := fmt.Sprintf("%s#%s", dsn, table)

	has_table, ok := lookup_table[lookup_key]

	if ok {
		return has_table, nil
	}

	check_tables := true
	has_table = false

	if !re_mem.MatchString(dsn) {

		test := dsn

		if re_file.MatchString(test) {

			s := re_file.FindAllStringSubmatch(dsn, -1)

			if len(s) == 1 && len(s[0]) == 2 {
				test = s[0][1]
			}
		}

		_, err := os.Stat(test)

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

	lookup_table[lookup_key] = has_table

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
