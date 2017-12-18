# go-whosonfirst-sqlite

Go package for working with Who's On First documents and SQLite databases.

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
	"github.com/whosonfirst/go-whosonfirst-sqlite/tables"
)

func main (){

	db, _ := database.NewDB("wof.db")
	defer db.Close()

	# Or you could just invoke these two calls with the handy:
	# st, _ := tables.NewSPRTableWithDatabase(db)

	st, _ := tables.NewSPRTable()
	st.InitializeTable(db)

	f, _ := feature.LoadWOFFeatureFromFile("123.geojson")
	st.IndexFeature(db, f)
}
```

_Error handling has been removed for the sake of brevity._

## Tables

### ancestors

```
CREATE TABLE ancestors (
	id INTEGER NOT NULL,
	ancestor_id INTEGER NOT NULL,
	ancestor_placetype TEXT,
	lastmodified INTEGER
);

CREATE INDEX ancestors_by_id ON ancestors (id,ancestor_placetype,lastmodified);
CREATE INDEX ancestors_by_ancestor ON ancestors (ancestor_id,ancestor_placetype,lastmodified);
CREATE INDEX ancestors_by_lastmod ON ancestors (lastmodified);
```

### geojson

```
CREATE TABLE geojson (
	id INTEGER NOT NULL PRIMARY KEY,
	body TEXT,
	lastmodified INTEGER
);

CREATE INDEX geojson_by_lastmod ON geojson (lastmodified);
```

### names

```
CREATE TABLE names (
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

CREATE INDEX names_by_lastmod ON names (lastmodified);
CREATE INDEX names_by_country ON names (country,privateuse,placetype);
CREATE INDEX names_by_language ON names (language,privateuse,placetype);
CREATE INDEX names_by_placetype ON names (placetype,country,privateuse);
CREATE INDEX names_by_name ON names (name, placetype, country);
CREATE INDEX names_by_name_private ON names (name, privateuse, placetype, country);
CREATE INDEX names_by_wofid ON names (id);
```

### spr

```
CREATE TABLE spr (
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

CREATE INDEX spr_by_lastmod ON spr (lastmodified);
CREATE INDEX spr_by_parent ON spr (parent_id, is_current, lastmodified);
CREATE INDEX spr_by_placetype ON spr (placetype, is_current, lastmodified);
CREATE INDEX spr_by_country ON spr (country, placetype, is_current, lastmodified);
CREATE INDEX spr_by_name ON spr (name, placetype, is_current, lastmodified);
CREATE INDEX spr_by_centroid ON spr (latitude, longitude, is_current, lastmodified);
CREATE INDEX spr_by_bbox ON spr (min_latitude, min_longitude, max_latitude, max_longitude, placetype, is_current, lastmodified);
CREATE INDEX spr_by_repo ON spr (repo, lastmodified);
CREATE INDEX spr_by_current ON spr (is_current, lastmodified);
CREATE INDEX spr_by_deprecated ON spr (is_deprecated, lastmodified);
CREATE INDEX spr_by_ceased ON spr (is_ceased, lastmodified);
CREATE INDEX spr_by_superseded ON spr (is_superseded, lastmodified);
CREATE INDEX spr_by_superseding ON spr (is_superseding, lastmodified);
```

## Custom tables

Sure. You just need to write a per-table package that implements the `Table` interface, described below.

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
     IndexFeature(Database, geojson.Feature) error
}
```

Where `geojson.Feature` is defined in the [go-whosonfirst-geojson-v2](https://github.com/whosonfirst/go-whosonfirst-geojson-v2#geojsonfeature) package.

## Tools

### wof-sqlite-index

```
./bin/wof-sqlite-index -h
Usage of ./bin/wof-sqlite-index:
  -all
    	Index all tables
  -ancestors
    	Index the 'ancestors' tables
  -dsn string
    	 (default ":memory:")
  -geojson
    	Index the 'geojson' table
  -live-hard-die-fast
    	Enable various performance-related pragmas at the expense of possible (unlikely) database corruption
  -mode string
    	The mode to use importing data. Valid modes are: directory,feature,feature-collection,files,geojson-ls,meta,path,repo. (default "files")
  -names
    	Index the 'names' table
  -processes int
    	The number of concurrent processes to index data with (default 16)
  -spr
    	Index the 'spr' table
  -timings
    	Display timings during and after indexing
```

For example:

```
./bin/wof-sqlite-index -live-hard-die-fast -dsn microhoods.db -all -mode meta /usr/local/data/whosonfirst-data/meta/wof-microhood-latest.csv
```

See the way we're passing a `-live-hard-die-fast` flag? That is to enable a number of [performace-related PRAGMA commands](https://blog.devart.com/increasing-sqlite-performance.html) without which database index can be prohibitive and time-consuming. These is a small but unlikely chance of database corruptions when this flag is enabled.

You can also use `wof-sqlite-index` in combination with the [go-whosonfirst-api](https://github.com/whosonfirst/go-whosonfirst-api) `wof-api` tool and populate your SQLite database by piping API results on STDIN. For example, here's how you might index all the neighbourhoods in Montreal:

```
/usr/local/bin/wof-api -param method=whosonfirst.places.getDescendants -param id=101736545 \
-param placetype=neighbourhood -param api_key=mapzen-xxxxxx -geojson-ls | \
/usr/local/bin/wof-sqlite-index -dsn neighbourhoods.db -all -mode geojson-ls STDIN
```

Or creating dated databases for all the placetypes in the `whosonfirst-data` repo. _It would be nice if this worked for other repos (like venues, postalcodes, etc) but it probably doesn't._

```
#!/bin/sh

REPO=$1

YMD=`date "+%Y%m%d"`

for META in `ls -a ${REPO}/meta/*.csv | grep -v concordances`
do
    FNAME=`basename ${META}`
    PLACETYPE=`echo ${FNAME} | awk -F '-' '{ print $2 }'`

    echo "make db for ${FNAME}"
    echo `date`
    ./bin/wof-sqlite-index -all -dsn ${PLACETYPE}-${YMD}.db -mode meta ${META}

done
```    

## See also

* https://whosonfirst.mapzen.com/sqlite
* https://sqlite.org/
