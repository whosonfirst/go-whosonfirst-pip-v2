# go-whosonfirst-sqlite-features

Go package for working with Who's On First features and SQLite databases.

## Install

You will need to have both `Go` (specifically a version of Go more recent than 1.6 so let's just assume you need [Go 1.8](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make bin
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

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

### concordances

```
CREATE TABLE concordances (
	id INTEGER NOT NULL,
	concordance_id INTEGER NOT NULL,
	concordance_souce TEXT,
	lastmodified INTEGER
);

CREATE INDEX concordances_by_id ON concordances (id,lastmodified);
CREATE INDEX concordances_by_other ON concordances (other_source,other_id);	
CREATE INDEX concordances_by_other_lastmod ON concordances (other_source,other_id,lastmodified);
CREATE INDEX ancestors_by_lastmod ON concordances (lastmodified);`
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

### geometries

```
CREATE TABLE geometries (
	id INTEGER NOT NULL PRIMARY KEY,
	is_alt TINYINT,
	type TEXT,
	lastmodified INTEGER
);

SELECT InitSpatialMetaData();
SELECT AddGeometryColumn('geometries', 'geom', 4326, 'GEOMETRY', 'XY');
SELECT CreateSpatialIndex('geometries', 'geom');

CREATE INDEX geometries_by_lastmod ON geometries (lastmodified);`
```

_Notes: In order to index geometries you will need to have the [Spatialite extension](https://www.gaia-gis.it/fossil/libspatialite/index) installed._

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

### search

```
CREATE VIRTUAL TABLE search USING fts4(
	id, placetype,
	name, names_all, names_preferred, names_variant, names_colloquial,		
	is_current, is_ceased, is_deprecated, is_superseded
);
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

Sure. You just need to write a per-table package that implements the `Table` interface as described in [go-whosonfirst-sqlite](https://github.com/whosonfirst/go-whosonfirst-sqlite#custom-tables).

## Tools

### wof-sqlite-index-features

```
./bin/wof-sqlite-index-features -h
Usage of ./bin/wof-sqlite-index-features:
  -all
    	Index all tables (except the 'search' and 'geometries' tables which you need to specify explicitly)
  -ancestors
    	Index the 'ancestors' tables
  -concordances
    	Index the 'concordances' tables
  -driver string
    	 (default "sqlite3")
  -dsn string
    	 (default ":memory:")
  -geojson
    	Index the 'geojson' table
  -geometries
    	Index the 'geometries' table (requires that libspatialite already be installed)
  -live-hard-die-fast
    	Enable various performance-related pragmas at the expense of possible (unlikely) database corruption
  -mode string
    	The mode to use importing data. Valid modes are: directory,feature,feature-collection,files,geojson-ls,meta,path,repo,sqlite. (default "files")
  -names
    	Index the 'names' table
  -processes int
    	The number of concurrent processes to index data with (default 16)
  -search
    	Index the 'search' table (using SQLite FTS4 full-text indexer)
  -spr
    	Index the 'spr' table
  -timings
    	Display timings during and after indexing
```

For example:

```
./bin/wof-sqlite-index-features -live-hard-die-fast -dsn microhoods.db -all -mode meta /usr/local/data/whosonfirst-data/meta/wof-microhood-latest.csv
```

See the way we're passing a `-live-hard-die-fast` flag? That is to enable a number of performace-related PRAGMA commands (described [here](https://blog.devart.com/increasing-sqlite-performance.html) and [here](https://www.gaia-gis.it/gaia-sins/spatialite-cookbook/html/system.html)) without which database index can be prohibitive and time-consuming. These is a small but unlikely chance of database corruptions when this flag is enabled.

Also note that the `-live-hard-die-fast` flag will cause the `PAGE_SIZE` and `CACHE_SIZE` PRAGMAs to be set to `4096` and `1000000` respectively so the eventual cache size will require 4GB of memory. This is probably fine on most systems where you'll be indexing data but I am open to the idea that we may need to revisit those numbers or at least make them configurable.

You can also use `wof-sqlite-index-features` in combination with the [go-whosonfirst-api](https://github.com/whosonfirst/go-whosonfirst-api) `wof-api` tool and populate your SQLite database by piping API results on STDIN. For example, here's how you might index all the neighbourhoods in Montreal:

```
/usr/local/bin/wof-api -param method=whosonfirst.places.getDescendants -param id=101736545 \
-param placetype=neighbourhood -param api_key=mapzen-xxxxxx -geojson-ls | \
/usr/local/bin/wof-sqlite-index-features -dsn neighbourhoods.db -all -mode geojson-ls STDIN
```

Or creating databases for all the Who's On First repos:

```
#!/bin/sh

for REPO in $@
do

    if [ ! -d ${REPO}/data ]
    then
	echo "${REPO} has no data directory"
	continue
    fi
    
    FNAME=`basename ${REPO}`
    echo "make db for ${FNAME}"

    if [ -f "/usr/local/data/whosonfirst-sqlite/${FNAME}.db" ]
    then
	rm /usr/local/data/whosonfirst-sqlite/${FNAME}.db
    fi

    ./bin/wof-sqlite-index-features -timings -live-hard-die-fast -all -dsn /usr/local/data/whosonfirst-sqlite/${FNAME}-latest.db -mode repo ${REPO} 

done
```    

### wof-sqlite-query-features

Query a search-enabled SQLite database by name(s). Results are output as CSV encoded rows containing `id` and `(wof:)name` properties.

_This assumes you have created the database using the `wof-sqlite-index-features` tool with the `-search` paramter._

```
./bin/wof-sqlite-query-features -h
Usage of ./bin/wof-sqlite-query-features:
  -column string
    	The 'names_*' column to query against. Valid columns are: names_all, names_preferred, names_variant, names_colloquial. (default "names_all")
  -driver string
    	 (default "sqlite3")
  -dsn string
    	 (default ":memory:")
  -is-ceased string
    	A comma-separated list of valid existential flags (-1,0,1) to filter results according to whether or not they have been marked as ceased. Multiple flags are evaluated as a nested 'OR' query.
  -is-current string
    	A comma-separated list of valid existential flags (-1,0,1) to filter results according to their 'mz:is_current' property. Multiple flags are evaluated as a nested 'OR' query.
  -is-deprecated string
    	A comma-separated list of valid existential flags (-1,0,1) to filter results according to whether or not they have been marked as deprecated. Multiple flags are evaluated as a nested 'OR' query.
  -is-superseded string
    	A comma-separated list of valid existential flags (-1,0,1) to filter results according to whether or not they have been marked as superseded. Multiple flags are evaluated as a nested 'OR' query.
  -output string
    	A valid path to write (CSV) results to. If empty results are written to STDOUT.
  -table string
    	The name of the SQLite table to query against. (default "search")
```

For example:

```
./bin/wof-sqlite-query-features -dsn test2.db JFK
102534365,John F Kennedy Int'l Airport

./bin/wof-sqlite-query-features -dsn test2.db -column names_colloquial Paris
85922583,San Francisco
102027181,Shanghai
102030585,Kolkata
101751929,Tromsø
```

Full-text search is supported using SQLite's FTS4 indexer. In order to index the `search` table you must explicitly pass the `-search` flag to the `wof-sqlite-index-features` command. It is _not_ included when you set the `-all` flag (which should probably be renamed to be `-common` but that's not the case today...) because it increases the overall indexing time by a non-trivial amount.

## Spatial indexes

Yes, if you have the [Spatialite extension](https://www.gaia-gis.it/fossil/libspatialite/index) installed and have indexed the `geometries` table. For example:

```
> ./bin/wof-sqlite-index-features -timings -live-hard-die-fast -spr -geometries -driver spatialite -mode repo -dsn test.db /usr/local/data/whosonfirst-data-constituency-ca/
10:09:46.534281 [wof-sqlite-index-features] STATUS time to index geometries (87) : 21.251828704s
10:09:46.534379 [wof-sqlite-index-features] STATUS time to index spr (87) : 3.206930799s
10:09:46.534385 [wof-sqlite-index-features] STATUS time to index all (87) : 24.48004637s

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

_Remember: When indexing geometries you will need to explcitly pass both the `-geometries` and `-driver spatialite` flags, even if you are already passing in the `-all` flag. This is so `-all` will continue to work as expected for people who don't have Spatialite installed on their computer._

## Indexing 

Indexing time will vary depending on the specifics of your hardware (available RAM, CPU, disk I/O) but as a rule building indexes with the `geometries` table will take longer, and create a larger database, than doing so without. For example indexing the [whosonfirst-data](https://github.com/whosonfirst-data/whosonfirst-data) repository with spatial indexes:

```
> ./bin/wof-sqlite-index-features -all -driver spatialite -geometries -dsn /usr/local/data/dist/sqlite/whosonfirst-data-latest.db -live-hard-die-fast -timings -mode repo /usr/local/data/whosonfirst-data
...time passes...
06:12:51.274132 [wof-sqlite-index-features] STATUS time to index geojson (951541) : 13m41.994217581s
06:12:51.274158 [wof-sqlite-index-features] STATUS time to index spr (951541) : 13m0.21007633s
06:12:51.274173 [wof-sqlite-index-features] STATUS time to index names (951541) : 17m50.759093941s
06:12:51.274178 [wof-sqlite-index-features] STATUS time to index ancestors (951541) : 3m37.431723948s
06:12:51.274182 [wof-sqlite-index-features] STATUS time to index concordances (951541) : 2m36.737857568s
06:12:51.274187 [wof-sqlite-index-features] STATUS time to index geometries (951541) : 43m48.39054903s
06:12:51.274192 [wof-sqlite-index-features] STATUS time to index all (951541) : 4h41m45.492361401s

> du -h /usr/local/data/dist/sqlite/whosonfirst-data-latest.db
15G     /usr/local/data/dist/sqlite/whosonfirst-data-latest.db
```

And without:

```
> ./bin/wof-sqlite-index-features -all -dsn /usr/local/data/dist/sqlite/whosonfirst-data-latest-nospatial.db -live-hard-die-fast -timings -mode repo /usr/local/data/whosonfirst-data
...time passes...
10:06:13.226187 [wof-sqlite-index-features] STATUS time to index names (951541) : 12m32.359733539s
10:06:13.226206 [wof-sqlite-index-features] STATUS time to index ancestors (951541) : 3m27.294843778s
10:06:13.226212 [wof-sqlite-index-features] STATUS time to index concordances (951541) : 2m5.947968206s
10:06:13.226220 [wof-sqlite-index-features] STATUS time to index geojson (951541) : 10m11.355455209s
10:06:13.226226 [wof-sqlite-index-features] STATUS time to index spr (951541) : 11m32.687081163s
10:06:13.226233 [wof-sqlite-index-features] STATUS time to index all (951541) : 3h43m20.687783762s

> du -h /usr/local/data/dist/sqlite/whosonfirst-data-latest-nospatial.db 
12G     /usr/local/data/dist/sqlite/whosonfirst-data-latest-nospatial.db
```

As of this writing individual tables are indexed atomically. There may be some improvements to be made indexing tables in separate Go routines but my hunch is this will make SQLite sad and cause a lot of table lock errors. I don't need to be right about that, though...

## See also

* https://sqlite.org/
* https://www.gaia-gis.it/fossil/libspatialite/index
* https://dist.whosonfirst.org/sqlite/
* https://github.com/whosonfirst/go-whosonfirst-sqlite
