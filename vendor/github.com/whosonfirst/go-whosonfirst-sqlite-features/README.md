# go-whosonfirst-sqlite-features

Go package for working with Who's On First features and SQLite databases.

## Install

You will need to have both `Go` (specifically version [Go 1.12](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make tools
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
	id INTEGER NOT NULL,
	body TEXT,
	source TEXT,
	is_alt BOOLEAN,
	alt_label TEXT,
	lastmodified INTEGER
);

CREATE UNIQUE INDEX geojson_by_id ON %s (id, source, alt_label);
CREATE INDEX geojson_by_alt ON %s (id, is_alt, alt_label);
CREATE INDEX geojson_by_lastmod ON %s (lastmodified);
```

### geometry

```
CREATE TABLE geometry (
	id INTEGER NOT NULL,
	body TEXT,
	is_alt BOOLEAN,
	alt_label TEXT,
	lastmodified INTEGER
);

CREATE UNIQUE INDEX geojson_by_id ON %s (id, alt_label);
CREATE INDEX geojson_by_alt ON %s (id, is_alt, alt_label);
CREATE INDEX geojson_by_lastmod ON %s (lastmodified);
```

This table indexes only the `geometry` elements for Who's On First records. This table is principally used in concert with the `rtree` for performance reasons.

### geometries

```
CREATE TABLE geometries (
	id INTEGER NOT NULL,
	type TEXT,
	is_alt TINYINT,
	alt_label TEXT,
	lastmodified INTEGER
);

SELECT InitSpatialMetaData();
SELECT AddGeometryColumn('geometries', 'geom', 4326, 'GEOMETRY', 'XY');
SELECT CreateSpatialIndex('geometries', 'geom');

CREATE UNIQUE INDEX by_id ON geometries (id, alt_label);
CREATE INDEX geometries_by_lastmod ON geometries (lastmodified);`
```

#### Notes

* In order to index the `geometries` table you will need to have the [Spatialite extension](https://www.gaia-gis.it/fossil/libspatialite/index) installed.
* As of Decemeber 2020, I am no longer able to make this (indexing the `geometries` table) work under OS X. I am not sure if this is a `spatialite` thing or a `go-sqlite3` thing or something else. Any help resolving this issue would be welcome.

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

### properties

```
CREATE TABLE properties (
	id INTEGER NOT NULL,
	body TEXT,
	is_alt BOOLEAN,
	alt_label TEXT,
	lastmodified INTEGER
);

CREATE UNIQUE INDEX geojson_by_id ON %s (id, alt_label);
CREATE INDEX geojson_by_alt ON %s (id, is_alt, alt_label);
CREATE INDEX geojson_by_lastmod ON %s (lastmodified);
```

This table indexes only the `properties` elements for Who's On First records. This table is principally used in concert with the `rtree` for performance reasons.

### rtree

```
CREATE VIRTUAL TABLE %s USING rtree (
	id,
	min_x,
	max_x,
	min_y,
	max_y,
	+wof_id INTEGER,
	+is_alt TINYINT,
	+alt_label TEXT,
	+geometry BLOB,	
	+lastmodified INTEGER
);
```

#### Notes

Section `3.1.1` of the [SQLite RTree documentation](https://www.sqlite.org/rtree.html) states:

> In the argments to "rtree" in the CREATE VIRTUAL TABLE statement, the names of the columns are taken from the first token of each argument. All subsequent tokens within each argument are silently ignored. This means, for example, that if you try to give a column a type affinity or add a constraint such as UNIQUE or NOT NULL or DEFAULT to a column, those extra tokens are accepted as valid, but they do not change the behavior of the rtree. In an RTREE virtual table, the first column always has a type affinity of INTEGER and all other data columns have a type affinity of NUMERIC. Recommended practice is to omit any extra tokens in the rtree specification. Let each argument to "rtree" be a single ordinary label that is the name of the corresponding column, and omit all other tokens from the argument list.

Section `4.1` goes on to say:

> Beginning with SQLite version 3.24.0 (2018-06-04), r-tree tables can have auxiliary columns that store arbitrary data. ... Auxiliary columns are marked with a "+" symbol before the column name. Auxiliary columns must come after all of the coordinate boundary columns. There is a limit of no more than 100 auxiliary columns.

With that in mind the `rtree` table relies on SQLite to automatically generate a new primary key value for the `id` column. The Who's On First record's ID is _not_ the primary key for the table and is stored in the `wof_id` column. It may be associated with (1) primary record and (n) alternate geometry records. If an alternate geometry is indexed the `is_alt` column value will be set to "1" and the `alt_label` column will be populated with the value in that record's `src:alt_label` property.

The `+geometry` column contains each polygons rings JSON-encoded as `[][][]float64`.

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
	is_alt TINYINT,
	alt_label TEXT,	
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
CREATE INDEX spr_obsolete ON spr (is_deprecated, is_superseded);
```

### supersedes

```
CREATE TABLE %s (
	id INTEGER NOT NULL,
	superseded_id INTEGER NOT NULL,
	superseded_by_id INTEGER NOT NULL,
	lastmodified INTEGER
);

CREATE UNIQUE INDEX supersedes_by ON %s (id,superseded_id, superseded_by_id);
```

## Custom tables

Sure. You just need to write a per-table package that implements the `Table` interface as described in [go-whosonfirst-sqlite](https://github.com/whosonfirst/go-whosonfirst-sqlite#custom-tables).

## Dependencies and relationships

These are documented in the [Dependencies and relationships section](https://github.com/whosonfirst/go-whosonfirst-sqlite#dependencies-and-relationships) of the `go-whosonfirst-sqlite` package.

## See also

* https://sqlite.org/
* https://www.gaia-gis.it/fossil/libspatialite/index
* https://github.com/whosonfirst/go-whosonfirst-sqlite
* https://github.com/whosonfirst/go-whosonfirst-sqlite-feature-index