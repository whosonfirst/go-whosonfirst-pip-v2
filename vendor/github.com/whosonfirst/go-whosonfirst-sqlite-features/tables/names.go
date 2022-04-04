package tables

import (
	"context"
	"fmt"
	"github.com/aaronland/go-sqlite"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-names/tags"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features"
)

type NamesTable struct {
	features.FeatureTable
	name string
}

type NamesRow struct {
	Id           int64
	Placetype    string
	Country      string
	Language     string
	ExtLang      string
	Script       string
	Region       string
	Variant      string
	Extension    string
	PrivateUse   string
	Name         string
	LastModified int64
}

func NewNamesTableWithDatabase(ctx context.Context, db sqlite.Database) (sqlite.Table, error) {

	t, err := NewNamesTable(ctx)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func NewNamesTable(ctx context.Context) (sqlite.Table, error) {

	t := NamesTable{
		name: "names",
	}

	return &t, nil
}

func (t *NamesTable) Name() string {
	return t.name
}

func (t *NamesTable) Schema() string {

	sql := `CREATE TABLE %s (
	       id INTEGER NOT NULL,
	       placetype TEXT,
	       country TEXT,
	       language TEXT,
	       extlang TEXT,
	       script TEXT,
	       region TEXT,
	       variant TEXT,
	       extension TEXT,
	       privateuse TEXT,
	       name TEXT,
	       lastmodified INTEGER
	);

	CREATE INDEX names_by_lastmod ON %s (lastmodified);
	CREATE INDEX names_by_country ON %s (country,privateuse,placetype);
	CREATE INDEX names_by_language ON %s (language,privateuse,placetype);
	CREATE INDEX names_by_placetype ON %s (placetype,country,privateuse);
	CREATE INDEX names_by_name ON %s (name, placetype, country);
	CREATE INDEX names_by_name_private ON %s (name, privateuse, placetype, country);
	CREATE INDEX names_by_wofid ON %s (id);
	`

	// this is a bit stupid really... (20170901/thisisaaronland)
	return fmt.Sprintf(sql, t.Name(), t.Name(), t.Name(), t.Name(), t.Name(), t.Name(), t.Name(), t.Name())
}

func (t *NamesTable) InitializeTable(ctx context.Context, db sqlite.Database) error {

	return sqlite.CreateTableIfNecessary(ctx, db, t)
}

func (t *NamesTable) IndexRecord(ctx context.Context, db sqlite.Database, i interface{}) error {
	return t.IndexFeature(ctx, db, i.(geojson.Feature))
}

func (t *NamesTable) IndexFeature(ctx context.Context, db sqlite.Database, f geojson.Feature) error {

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

	pt := f.Placetype()
	co := whosonfirst.Country(f)

	lastmod := whosonfirst.LastModified(f)
	names := whosonfirst.Names(f)

	for tag, names := range names {

		lt, err := tags.NewLangTag(tag)

		if err != nil {
			return err
		}

		for _, n := range names {

			if err != nil {
				return err
			}

			sql := fmt.Sprintf(`INSERT INTO %s (
	    			id, placetype, country,
				language, extlang,
				region, script, variant,
	    			extension, privateuse,
				name,
	    			lastmodified
			) VALUES (
	    		  	?, ?, ?,
				?, ?,
				?, ?, ?,
				?, ?,
				?,
				?
			)`, t.Name())

			stmt, err := tx.Prepare(sql)

			if err != nil {
				return err
			}

			defer stmt.Close()

			_, err = stmt.Exec(id, pt, co, lt.Language(), lt.ExtLang(), lt.Script(), lt.Region(), lt.Variant(), lt.Extension(), lt.PrivateUse(), n, lastmod)

			if err != nil {
				return err
			}

		}
	}

	return tx.Commit()
}
