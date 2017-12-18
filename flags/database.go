package flags

import (
	"errors"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"github.com/whosonfirst/go-whosonfirst-sqlite/utils"
	"strings"
)

type Database []string

func (e *Database) String() string {
	return strings.Join(*e, "\n")
}

func (e *Database) Set(path string) error {

	db, err := database.NewDB(path)

	if err != nil {
		return err
	}

	defer db.Close()

	has_table, err := utils.HasTable(db, "geojson")

	if err != nil {
		return err
	}

	if !has_table {
		return errors.New("Missing geojson table")
	}

	*e = append(*e, path)
	return nil
}
