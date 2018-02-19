package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"github.com/whosonfirst/go-whosonfirst-sqlite/utils"
	"io"
	"os"
	"strings"
)

func main() {

	driver := flag.String("driver", "sqlite3", "")
	var dsn = flag.String("dsn", ":memory:", "")
	var is_current = flag.String("is-current", "", "A comma-separated list of valid existential flags (-1,0,1) to filter results according to their 'mz:is_current' property. Multiple flags are evaluated as a nested 'OR' query.")
	var is_ceased = flag.String("is-ceased", "", "A comma-separated list of valid existential flags (-1,0,1) to filter results according to whether or not they have been marked as ceased. Multiple flags are evaluated as a nested 'OR' query.")
	var is_deprecated = flag.String("is-deprecated", "", "A comma-separated list of valid existential flags (-1,0,1) to filter results according to whether or not they have been marked as deprecated. Multiple flags are evaluated as a nested 'OR' query.")
	var is_superseded = flag.String("is-superseded", "", "A comma-separated list of valid existential flags (-1,0,1) to filter results according to whether or not they have been marked as superseded. Multiple flags are evaluated as a nested 'OR' query.")

	var table = flag.String("table", "search", "The name of the SQLite table to query against.")
	var col = flag.String("column", "names_all", "The 'names_*' column to query against. Valid columns are: names_all, names_preferred, names_variant, names_colloquial.")

	var output = flag.String("output", "", "A valid path to write (CSV) results to. If empty results are written to STDOUT.")

	flag.Parse()

	logger := log.SimpleWOFLogger()

	stdout := io.Writer(os.Stdout)
	logger.AddLogger(stdout, "status")

	var out io.Writer

	if *output == "" {
		out = os.Stdout
	} else {

		fh, err := os.OpenFile(*output, os.O_RDWR|os.O_CREATE, 0644)

		if err != nil {
			logger.Fatal("Unable to open file %s, because %s", *output, err)
		}

		out = fh
	}

	db, err := database.NewDBWithDriver(*driver, *dsn)

	if err != nil {
		logger.Fatal("Unable to create database (%s) because %s", *dsn, err)
	}

	defer db.Close()

	conn, err := db.Conn()

	if err != nil {
		logger.Fatal("Failed to connect to database, because %s", err)
	}

	match := fmt.Sprintf("%s MATCH ?", *col)
	query := strings.Join(flag.Args(), " ")

	conditions := []string{
		match,
	}

	args := []interface{}{
		query,
	}

	existential := map[string]string{
		"is_current":    *is_current,
		"is_ceased":     *is_ceased,
		"is_deprecated": *is_deprecated,
		"is_superseded": *is_superseded,
	}

	for label, flags := range existential {

		if flags == "" {
			continue
		}

		fl_conditions, fl_args, err := utils.ExistentialFlagsToQueryConditions(label, flags)

		if err != nil {
			logger.Fatal("Invalid '%s' flags (%s) %s", label, flags, err)
		}

		conditions = append(conditions, fl_conditions)

		for _, a := range fl_args {
			args = append(args, a)
		}
	}

	where := strings.Join(conditions, " AND ")

	sql := fmt.Sprintf("SELECT id,name FROM %s WHERE %s", *table, where)
	rows, err := conn.Query(sql, args...)

	if err != nil {
		logger.Fatal("Failed to query database (%s) because %s", sql, err)
	}

	defer rows.Close()

	wr := csv.NewWriter(out)

	for rows.Next() {

		var id string
		var name string

		err = rows.Scan(&id, &name)

		if err != nil {
			logger.Fatal("Failed to scan results, because %s", err)
		}

		row := []string{id, name}
		err := wr.Write(row)

		if err != nil {
			logger.Fatal("Failed to write CSV row because %s", err)
		}
	}

	err = rows.Err()

	if err != nil {
		logger.Fatal("The database is unhappy, because %s", err)
	}

	wr.Flush()

	err = wr.Error()

	if err != nil {
		logger.Fatal("The CSV writer is unhappy, because %s", err)
	}

	os.Exit(0)
}
