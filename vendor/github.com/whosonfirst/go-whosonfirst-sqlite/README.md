# go-whosonfirst-sqlite

Go package for working with SQLite databases.

## Install

You will need to have both `Go` (specifically a version of Go more recent than 1.6 so let's just assume you need [Go 1.8](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make bin
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Example

### Simple

```
import (
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
)

func main (){

	db, _ := database.NewDB("wof.db")
	defer db.Close()

	# Or you could just invoke these two calls with the handy:
	# st, _ := tables.NewSPRTableWithDatabase(db)

	st, _ := features.NewSPRTable()
	st.InitializeTable(db)

	f, _ := feature.LoadWOFFeatureFromFile("123.geojson")
	st.IndexFeature(db, f)
}
```

_Error handling has been removed for the sake of brevity._

## Tables

_If you're looking for all the tables related to Who's On First documents they've been moved in to the [go-whosonfirst-sqlite-features](https://github.com/whosonfirst/go-whosonfirst-sqlite-features) package._

## Custom tables

Sure. You just need to write a per-table package that implements the `Table` interface, described below. For examples, consult the `tables` directories in the [go-whosonfirst-sqlite-features](https://github.com/whosonfirst/go-whosonfirst-sqlite-features) or [go-whosonfirst-sqlite-brands](https://github.com/whosonfirst/go-whosonfirst-sqlite-brands) packages.

## DSN strings

### :memory:

To account for [this issue](https://github.com/mattn/go-sqlite3/issues/204) DSN strings that are `:memory:` will be rewritten as:

`file::memory:?mode=memory&cache=shared`

### things that don't start with `file:`

To account for [this issue](https://github.com/mattn/go-sqlite3/issues/39) DSN strings that are _not_ `:memory:` and _don't_ start with `:file:` will be rewritten as:

`file:{DSN}?cache=shared&mode=rwc`

## Interfaces

### Database

```
type Database interface {
     Conn() (*sql.DB, error)
     DSN() string
     Close() error
}
```

### Table

```
type Table interface {
     Name() string
     Schema() string
     InitializeTable(Database) error
     IndexRecord(Database, interface{}) error
}
```

It is left up to people implementing the `Table` interface to figure out what to do with the second value passed to the `IndexRecord` method. For example:

```
func (t *BrandsTable) IndexRecord(db sqlite.Database, i interface{}) error {
	return t.IndexBrand(db, i.(brands.Brand))
}

func (t *BrandsTable) IndexBrand(db sqlite.Database, b brands.Brand) error {
	// code to index brands.Brands here
}
```

## Spatial indexes

Yes, if you have the [Spatialite extension](https://www.gaia-gis.it/fossil/libspatialite/index) installed and have indexed the `geometries` table. For example:

```
> ./bin/wof-sqlite-index-features -timings -live-hard-die-fast -spr -geometries -driver spatialite -mode repo -dsn test.db /usr/local/data/whosonfirst-data-constituency-ca/
10:09:46.534281 [wof-sqlite-index] STATUS time to index geometries (87) : 21.251828704s
10:09:46.534379 [wof-sqlite-index] STATUS time to index spr (87) : 3.206930799s
10:09:46.534385 [wof-sqlite-index] STATUS time to index all (87) : 24.48004637s

> sqlite3 test.db
SQLite version 3.21.0 2017-10-24 18:55:49
Enter ".help" for usage hints.

sqlite> SELECT load_extension('mod_spatialite.dylib');
sqlite> SELECT s.id, s.name FROM spr s, geometries g WHERE ST_Intersects(g.geom, GeomFromText('POINT(-122.229137 49.450129)', 4326)) AND g.id = s.id;
1108962831|Maple Ridge-Pitt Meadows
```

Or:

```
> spatialite whosonfirst-data-latest.db
SpatiaLite version ..: 4.1.1	Supported Extensions:
...spatialite chatter goes here...
SQLite version 3.8.2 2013-12-06 14:53:30
Enter ".help" for instructions
Enter SQL statements terminated with a ";

spatialite> SELECT s.id, s.name FROM spr AS s, geometries AS g1, geometries AS g2 WHERE g1.id =  85834637 AND s.placetype = 'neighbourhood' AND g2.id = s.id AND ST_Touches(g1.geom, g2.geom) AND g2.ROWID IN (SELECT ROWID FROM SpatialIndex WHERE f_table_name = 'geometries' AND search_frame=g2.geom);
102112179|La Lengua
1108831803|Showplace Square

spatialite> SELECT s.id, s.name FROM spr AS s, geometries AS g1, geometries AS g2 WHERE g1.id != g2.id AND g1.id =  85865959 AND s.placetype = 'neighbourhood' AND s.is_current=1 AND g2.id = s.id AND (ST_Touches(g1.geom, g2.geom) OR ST_Intersects(g1.geom, g2.geom)) AND g2.ROWID IN (SELECT ROWID FROM SpatialIndex WHERE f_table_name = 'geometries' AND search_frame=g2.geom);
1108831807|Fairmount
85814471|Diamond Heights
85869221|Eureka Valley

SELECT s.id, s.name, s.is_current FROM spr AS s, geometries AS g1, geometries AS g2 WHERE g1.id != g2.id AND g1.id =  102061079 AND s.placetype = 'neighbourhood' AND g2.id = s.id AND (ST_Touches(g1.geom, g2.geom) OR ST_Intersects(g1.geom, g2.geom)) AND g2.ROWID IN (SELECT ROWID FROM SpatialIndex WHERE f_table_name = 'geometries' AND search_frame=g2.geom);
85892915|BoCoCa|0
85869125|Boerum Hill|1
420782915|Carroll Gardens|1
85865587|Gowanus|1
```

## Indexing 

Indexing time will vary depending on the specifics of your hardware (available RAM, CPU, disk I/O) but as a rule building indexes with the `geometries` table will take longer, and create a larger database, than doing so without. For example indexing the [whosonfirst-data](https://github.com/whosonfirst-data/whosonfirst-data) repository with spatial indexes:

```
> ./bin/wof-sqlite-index-features -all -driver spatialite -geometries -dsn /usr/local/data/dist/sqlite/whosonfirst-data-latest.db -live-hard-die-fast -timings -mode repo /usr/local/data/whosonfirst-data
...time passes...
06:12:51.274132 [wof-sqlite-index] STATUS time to index geojson (951541) : 13m41.994217581s
06:12:51.274158 [wof-sqlite-index] STATUS time to index spr (951541) : 13m0.21007633s
06:12:51.274173 [wof-sqlite-index] STATUS time to index names (951541) : 17m50.759093941s
06:12:51.274178 [wof-sqlite-index] STATUS time to index ancestors (951541) : 3m37.431723948s
06:12:51.274182 [wof-sqlite-index] STATUS time to index concordances (951541) : 2m36.737857568s
06:12:51.274187 [wof-sqlite-index] STATUS time to index geometries (951541) : 43m48.39054903s
06:12:51.274192 [wof-sqlite-index] STATUS time to index all (951541) : 4h41m45.492361401s

> du -h /usr/local/data/dist/sqlite/whosonfirst-data-latest.db
15G     /usr/local/data/dist/sqlite/whosonfirst-data-latest.db
```

And without:

```
> ./bin/wof-sqlite-index-features -all -dsn /usr/local/data/dist/sqlite/whosonfirst-data-latest-nospatial.db -live-hard-die-fast -timings -mode repo /usr/local/data/whosonfirst-data
...time passes...
10:06:13.226187 [wof-sqlite-index] STATUS time to index names (951541) : 12m32.359733539s
10:06:13.226206 [wof-sqlite-index] STATUS time to index ancestors (951541) : 3m27.294843778s
10:06:13.226212 [wof-sqlite-index] STATUS time to index concordances (951541) : 2m5.947968206s
10:06:13.226220 [wof-sqlite-index] STATUS time to index geojson (951541) : 10m11.355455209s
10:06:13.226226 [wof-sqlite-index] STATUS time to index spr (951541) : 11m32.687081163s
10:06:13.226233 [wof-sqlite-index] STATUS time to index all (951541) : 3h43m20.687783762s

> du -h /usr/local/data/dist/sqlite/whosonfirst-data-latest-nospatial.db 
12G     /usr/local/data/dist/sqlite/whosonfirst-data-latest-nospatial.db
```

As of this writing individual tables are indexed atomically. There may be some improvements to be made indexing tables in separate Go routines but my hunch is this will make SQLite sad and cause a lot of table lock errors. I don't need to be right about that, though...

## See also

* https://sqlite.org/
* https://www.gaia-gis.it/fossil/libspatialite/index
* https://dist.whosonfirst.org/sqlite/
