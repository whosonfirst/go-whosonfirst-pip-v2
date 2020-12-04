package tables

import (
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features"
	"github.com/whosonfirst/go-whosonfirst-sqlite/utils"
)

type SPRTable struct {
	features.FeatureTable
	name string
}

type SPRRow struct {
	Id            int64   // properties.wof:id	INTEGER
	ParentId      int64   // properties.wof:parent_id	INTEGER
	Name          string  // properties.wof:name  TEXT
	Placetype     string  // properties.wof:placetype TEXT
	Country       string  // properties.wof:country TEXT
	Repo          string  // properties.wof:repo TEXT
	Path          string  // derived TEXT
	URI           string  // derived TEXT
	Latitude      float64 // derived REAL
	Longitude     float64 // derived REAL
	MinLatitude   float64 // properties.geom:bbox.1 REAL
	MinLongitude  float64 // properties.geom:bbox.0 REAL
	MaxLatitude   float64 // properties.geom:bbox.3 REAL
	MaxLongitude  float64 // properties.geom.bbox.2 REAL
	IsCurrent     int64   // properies.mz:is_current INTEGER
	IsCeased      int64   // derived INTEGER
	IsDeprecated  int64   // derived INTEGER
	IsSuperseded  int64   // derived INTEGER
	IsSuperseding int64   // derived INTEGER
	SupersededBy  []int64 // ...
	Supersedes    []int64 // ...
	LastModified  int64   // properties.wof:lastmodified INTEGER
}

func NewSPRTableWithDatabase(db sqlite.Database) (sqlite.Table, error) {

	t, err := NewSPRTable()

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func NewSPRTable() (sqlite.Table, error) {

	t := SPRTable{
		name: "spr",
	}

	return &t, nil
}

func (t *SPRTable) InitializeTable(db sqlite.Database) error {

	return utils.CreateTableIfNecessary(db, t)
}

func (t *SPRTable) Name() string {
	return t.name
}

func (t *SPRTable) Schema() string {

	sql := `CREATE TABLE %[1]s (
			id INTEGER NOT NULL PRIMARY KEY,
			parent_id INTEGER,
			name TEXT,
			placetype TEXT,
			country TEXT,
			repo TEXT,
			latitude REAL,
			longitude REAL,
			min_latitude REAL,
			min_longitude REAL,
			max_latitude REAL,
			max_longitude REAL,
			is_current INTEGER,
			is_deprecated INTEGER,
			is_ceased INTEGER,
			is_superseded INTEGER,
			is_superseding INTEGER,
			superseded_by TEXT,
			supersedes TEXT,
			lastmodified INTEGER
	);

	CREATE INDEX spr_by_lastmod ON %[1]s (lastmodified);
	CREATE INDEX spr_by_parent ON %[1]s (parent_id, is_current, lastmodified);
	CREATE INDEX spr_by_placetype ON %[1]s (placetype, is_current, lastmodified);
	CREATE INDEX spr_by_country ON %[1]s (country, placetype, is_current, lastmodified);
	CREATE INDEX spr_by_name ON %[1]s (name, placetype, is_current, lastmodified);
	CREATE INDEX spr_by_centroid ON %[1]s (latitude, longitude, is_current, lastmodified);
	CREATE INDEX spr_by_bbox ON %[1]s (min_latitude, min_longitude, max_latitude, max_longitude, placetype, is_current, lastmodified);
	CREATE INDEX spr_by_repo ON %[1]s (repo, lastmodified);
	CREATE INDEX spr_by_current ON %[1]s (is_current, lastmodified);
	CREATE INDEX spr_by_deprecated ON %[1]s (is_deprecated, lastmodified);
	CREATE INDEX spr_by_ceased ON %[1]s (is_ceased, lastmodified);
	CREATE INDEX spr_by_superseded ON %[1]s (is_superseded, lastmodified);
	CREATE INDEX spr_by_superseding ON %[1]s (is_superseding, lastmodified);
	CREATE INDEX spr_obsolete ON %[1]s (is_deprecated, is_superseded);
	`

	return fmt.Sprintf(sql, t.Name())
}

func (t *SPRTable) IndexRecord(db sqlite.Database, i interface{}) error {
	return t.IndexFeature(db, i.(geojson.Feature))
}

func (t *SPRTable) IndexFeature(db sqlite.Database, f geojson.Feature) error {

	is_alt := whosonfirst.IsAlt(f)

	if is_alt {
		return nil
	}

	spr, err := f.SPR()

	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
		id, parent_id, name, placetype,
		country, repo,
		latitude, longitude,
		min_latitude, min_longitude,
		max_latitude, max_longitude,
		is_current, is_deprecated, is_ceased,
		is_superseded, is_superseding,
		superseded_by, supersedes,
		lastmodified
		) VALUES (
		?, ?, ?, ?,
		?, ?,
		?, ?,
		?, ?,
		?, ?,
		?, ?, ?,
		?, ?,
		?, ?,
		?
		)`, t.Name()) // ON CONFLICT DO BLAH BLAH BLAH

	args := []interface{}{
		spr.Id(), spr.ParentId(), spr.Name(), spr.Placetype(),
		spr.Country(), spr.Repo(),
		spr.Latitude(), spr.Longitude(),
		spr.MinLatitude(), spr.MinLongitude(),
		spr.MaxLatitude(), spr.MaxLongitude(),
		spr.IsCurrent().Flag(), spr.IsDeprecated().Flag(), spr.IsCeased().Flag(),
		spr.IsSuperseded().Flag(), spr.IsSuperseding().Flag(),
		"", "",
		spr.LastModified(),
	}

	conn, err := db.Conn()

	if err != nil {
		return err
	}

	tx, err := conn.Begin()

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
