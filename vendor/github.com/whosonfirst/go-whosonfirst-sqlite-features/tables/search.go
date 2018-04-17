package tables

import (
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-names/tags"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features"
	"github.com/whosonfirst/go-whosonfirst-sqlite/utils"
	_ "log"
	"strings"
)

type SearchTable struct {
	features.FeatureTable
	name string
}

func NewSearchTableWithDatabase(db sqlite.Database) (sqlite.Table, error) {

	t, err := NewSearchTable()

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func NewSearchTable() (sqlite.Table, error) {

	t := SearchTable{
		name: "search",
	}

	return &t, nil
}

func (t *SearchTable) InitializeTable(db sqlite.Database) error {

	return utils.CreateTableIfNecessary(db, t)
}

func (t *SearchTable) Name() string {
	return t.name
}

func (t *SearchTable) Schema() string {

	schema := `CREATE VIRTUAL TABLE %s USING fts4(
		id, placetype,
		name, names_all, names_preferred, names_variant, names_colloquial,		
		is_current, is_ceased, is_deprecated, is_superseded
	);`

	// so dumb...
	return fmt.Sprintf(schema, t.Name())
}

func (t *SearchTable) IndexRecord(db sqlite.Database, i interface{}) error {
	return t.IndexFeature(db, i.(geojson.Feature))
}

func (t *SearchTable) IndexFeature(db sqlite.Database, f geojson.Feature) error {

	is_current, err := whosonfirst.IsCurrent(f)

	if err != nil {
		return err
	}

	is_ceased, err := whosonfirst.IsCeased(f)

	if err != nil {
		return err
	}

	is_deprecated, err := whosonfirst.IsDeprecated(f)

	if err != nil {
		return err
	}

	is_superseded, err := whosonfirst.IsSuperseded(f)

	if err != nil {
		return err
	}

	names_all := make([]string, 0)
	names_preferred := make([]string, 0)
	names_variant := make([]string, 0)
	names_colloquial := make([]string, 0)

	for tag, names := range whosonfirst.Names(f) {

		lt, err := tags.NewLangTag(tag)

		if err != nil {
			return err
		}

		possible := make([]string, 0)
		possible_map := make(map[string]bool)

		for _, n := range names {

			_, ok := possible_map[n]

			if !ok {
				possible_map[n] = true
			}
		}

		for n, _ := range possible_map {
			possible = append(possible, n)
		}

		for _, n := range possible {
			names_all = append(names_all, n)
		}

		switch lt.PrivateUse() {
		case "x_preferred":
			for _, n := range possible {
				names_preferred = append(names_preferred, n)
			}
		case "x_variant":
			for _, n := range possible {
				names_variant = append(names_variant, n)
			}
		case "x_colloquial":
			for _, n := range possible {
				names_colloquial = append(names_colloquial, n)
			}
		default:
			continue
		}
	}

	sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
		id, placetype,
		name, names_all, names_preferred, names_variant, names_colloquial,		
		is_current, is_ceased, is_deprecated, is_superseded
		) VALUES (
		?, ?,
		?, ?, ?, ?, ?,
		?, ?, ?, ?
		)`, t.Name()) // ON CONFLICT DO BLAH BLAH BLAH

	args := []interface{}{
		f.Id(), f.Placetype(),
		f.Name(), strings.Join(names_all, " "), strings.Join(names_preferred, " "), strings.Join(names_variant, " "), strings.Join(names_colloquial, " "),
		is_current.Flag(), is_ceased.Flag(), is_deprecated.Flag(), is_superseded.Flag(),
	}

	conn, err := db.Conn()

	if err != nil {
		return err
	}

	tx, err := conn.Begin()

	if err != nil {
		return err
	}

	s, err := tx.Prepare(fmt.Sprintf("DELETE FROM %s WHERE id = ?", t.Name()))

	if err != nil {
		return err
	}

	defer s.Close()

	_, err = s.Exec(f.Id())

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(args...)

	if err != nil {
		return err
	}

	return tx.Commit()
}
