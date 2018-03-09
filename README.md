# go-whosonfirst-pip-v2

An in-memory point-in-polygon (reverse geocoding) package for GeoJSON data, principally Who's On First data.

_This package supersedes the [go-whosonfirst-pip](https://github.com/whosonfirst/go-whosonfirst-pip) package which is no longer maintained._

## Install

You will need to have both `Go` (specifically a version of Go more recent than 1.7 so let's just assume you need [Go 1.9](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make bin
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Important

This package lacks normal Go documentation for packages and methods. Also normal
Go tests. Both are on the list (and contributions for either would be welcome)
but in the meantime there is documentation (-ish) below.

## Differences from "v1"

Pretty much everything under the hood has changed as have the public interfaces
since the [first
release](https://github.com/whosonfirst/go-whosonfirst-geojson-v2/blob/master/feature/whosonfirst.go)
(so called "v1") of this package. In _broad stokes_ both packages have the same goal and do the same
thing. The main differences between the two packages are: 

* Decoupling of the indexing layers (to allow for alternatives to the default RTree implementation) and the caching layers and making proper interfaces for both
* The ability to filter results by placetype or existential flags (is current, is deprecated, etc.)
* The use of the [go-whosonfirst-geojson-v2](https://github.com/whosonfirst/go-whosonfirst-geojson-v2) package for working with Who's On First documents.
* The use of the [go-whosonfirst-spr](https://github.com/whosonfirst/go-whosonfirst-spr) package (SPR is an acronym for "standard place response") for handling and generating responses. 
* The use of the [go-whosonfirst-flags](https://github.com/whosonfirst/go-whosonfirst-flags) package for filtering results. 

## Example

### wof-pip-server

To run as an HTTP based point-in-polygon (PIP) server indexing Who's On First documents from a local disk, specify "-mode directory" and give the data directory as the first (non optional) argument.

```
./bin/wof-pip-server -mode directory /usr/local/data/whosonfirst-data/data
12:25:55.267986 [wof-pip-server] STATUS listening on localhost:8080
12:25:56.272296 [wof-pip-server] STATUS indexing 3023 records indexed
12:25:57.271904 [wof-pip-server] STATUS indexing 6554 records indexed
12:25:58.273400 [wof-pip-server] STATUS indexing 10200 records indexed
12:25:59.276565 [wof-pip-server] STATUS indexing 13596 records indexed
...
12:34:37.980572 [wof-pip-server] STATUS finished indexing
```

_You can index any valid "mode" as defined by the [go-whosonfirst-index](https://github.com/whosonfirst/go-whosonfirst-index#modes) package._

Your PIP server will now be answering requests on `localhost:8000`. For example:

```
curl -s 'http://localhost:8000/?latitude=37.794906&longitude=-122.395229&placetype=microhood' | python -mjson.tool
{
    "places": [
        {
            "mz:is_ceased": 1,
            "mz:is_current": 0,
            "mz:is_deprecated": 0,
            "mz:is_superseded": 0,
            "mz:is_superseding": 0,
            "mz:latitude": 37.794906,
            "mz:longitude": -122.395229,
            "mz:max_latitude": 37.796684756991,
            "mz:max_longitude": -122.39310801029,
            "mz:min_latitude": 37.792339744389,
            "mz:min_longitude": -122.39753901958,
            "mz:uri": "https://data.whosonfirst.org/420/561/633/420561633.geojson",
            "wof:country": "US",
            "wof:id": 420561633,
            "wof:lastmodified": 1501284302,
            "wof:name": "Super Bowl City",
            "wof:parent_id": 85865899,
            "wof:path": "420/561/633/420561633.geojson",
            "wof:placetype": "microhood",
            "wof:repo": "whosonfirst-data",
            "wof:superseded_by": [],
            "wof:supersedes": []
        }
    ]
}
```

Detailed documentation for `wof-pip-server` is included below.

## Responses

The default response format is a [standard places
result](https://www.whosonfirst.org/docs/spr/) (SPR) and more
specifically something that implements the `spr.StandardPlacesResults`
interface, so really just a list of SPRs.

Under the hood this package uses the
[go-whosonfirst-geojson-v2](https://github.com/whosonfirst/go-whosonfirst-geojson-v2)
package for working with GeoJSON documents. In order to accomodate various Who's
On First -isms that package has two separate GeoJSON parser thingies (one for
Who's On First GeoJSON and one for everything) each of which implements the
`SPR` interface but with different serializations.

The [Who's On
First](https://github.com/whosonfirst/go-whosonfirst-geojson-v2/blob/master/feature/whosonfirst.go)
SPR looks like this:

```
type WOFStandardPlacesResult struct {
	spr.StandardPlacesResult `json:",omitempty"`
	WOFId                    int64   `json:"wof:id"`
	WOFParentId              int64   `json:"wof:parent_id"`
	WOFName                  string  `json:"wof:name"`
	WOFPlacetype             string  `json:"wof:placetype"`
	WOFCountry               string  `json:"wof:country"`
	WOFRepo                  string  `json:"wof:repo"`
	WOFPath                  string  `json:"wof:path"`
	WOFSupersededBy          []int64 `json:"wof:superseded_by"`
	WOFSupersedes            []int64 `json:"wof:supersedes"`
	MZURI                    string  `json:"mz:uri"`
	MZLatitude               float64 `json:"mz:latitude"`
	MZLongitude              float64 `json:"mz:longitude"`
	MZMinLatitude            float64 `json:"mz:min_latitude"`
	MZMinLongitude           float64 `json:"mz:min_longitude"`
	MZMaxLatitude            float64 `json:"mz:max_latitude"`
	MZMaxLongitude           float64 `json:"mz:max_longitude"`
	MZIsCurrent              int64   `json:"mz:is_current"`
	MZIsCeased               int64   `json:"mz:is_ceased"`
	MZIsDeprecated           int64   `json:"mz:is_deprecated"`
	MZIsSuperseded           int64   `json:"mz:is_superseded"`
	MZIsSuperseding          int64   `json:"mz:is_superseding"`
	WOFLastModified          int64   `json:"wof:lastmodified"`
}
```

The [generic GeoJSON
](https://github.com/whosonfirst/go-whosonfirst-geojson-v2/blob/master/feature/geojson.go) SPR looks like this:

```
type GeoJSONStandardPlacesResult struct {
     spr.StandardPlacesResult `json:",omitempty"`
     SPRId                    string  `json:"spr:id"`
     SPRName                  string  `json:"spr:name"`
     SPRPlacetype             string  `json:"spr:placetype"`
     SPRLatitude              float64 `json:"spr:latitude"`
     SPRLongitude             float64 `json:"spr:longitude"`
     SPRMinLatitude           float64 `json:"spr:min_latitude"`
     SPRMinLongitude          float64 `json:"spr:min_longitude"`
     SPRMaxLatitude           float64 `json:"spr:max_latitude"`
     SPRMaxLongitude          float64 `json:"spr:max_longitude"`
}
```

It is also possible to request GeoJSON formatted responses either by calling the
[utils.ResultsToFeatureCollection() method](https://github.com/whosonfirst/go-whosonfirst-pip-v2/blob/master/utils/utils.go) in code or by passing in a
`?format=geojson` flag in an HTTP request (assuming that `wof-pip-server` has
been started with the `-enable-geojson` flag).

### Extras

It is possible to append custom _extra_ parameters to responses with the use of
a custom "extras" SQLite database. This work has not been formalized yet (like
does it deserve to have a proper interface or a separate standalone package) and
should still be considered experimental.

As of this writing extras are only supported by the `wof-pip-server` tool and
need to be invoked with the `-enable-extras` flag. The default DSN for the
extras database (as defined by the `-extras-dsn` flag) is `:tmpfile:` which
means that a temporary SQLite database will be created and populated at index
time and then deleted when the program exits.

To query for extras (when calling the `wof-pip-server`) simply pass along a
comma-separated list of strings to the `extras` parameter. For example:

```
// ./bin/wof-pip-server -index spatialite -cache spatialite -spatialite-dsn \
//   /usr/local/data/whosonfirst-data-constituency-us-latest.db -enable-www \
//   -enable-extras -extras-dsn /usr/local/data/whosonfirst-data-constituency-us-latest.db \
//   -mode spatialite

curl 'http://localhost:8080/?latitude=37.6588&longitude=-122.4979&extras=geom:'

{
  "places": [
    {
      "geom:area": 0.152975,
      "geom:area_square_m": 1499191981.266914,
      "geom:bbox": "-123.173825,37.311653,-122.081473,37.823058",
      "geom:latitude": 37.573675,
      "geom:longitude": -122.495153,
      "mz:is_ceased": -1,
      "mz:is_current": -1,
      "mz:is_deprecated": 0,
      "mz:is_superseded": 0,
      "mz:is_superseding": 0,
      "mz:latitude": 37.573675,
      "mz:longitude": -122.495153,
      "mz:max_latitude": 37.823058,
      "mz:max_longitude": -122.081473,
      "mz:min_latitude": 37.311653,
      "mz:min_longitude": -123.173825,
      "mz:uri": "https://data.whosonfirst.org/110/873/834/7/1108738347.geojson",
      "wof:country": "us",
      "wof:id": 1108738347,
      "wof:lastmodified": 1493955495,
      "wof:name": "California Congressional District 14",
      "wof:parent_id": 85688637,
      "wof:path": "110/873/834/7/1108738347.geojson",
      "wof:placetype": "constituency",
      "wof:repo": "whosonfirst-data-constituency-us",
      "wof:superseded_by": [],
      "wof:supersedes": []
    },
    ... and so on
   ]
}
```

Extras themselves can be defined as fully-qualified keys or use a wildcard
notation of `{PREFIX}:` or `{PREFIX}:*` to retrieve all the keys matching a
given prefix.

The code to append extras is defined in the [extras package](extras/extras.go)
and the first thing to understand is that _it operates on raw JSON bytes_ rather
than a strictly defined interface like the SPR.

The basic signature for appending extras is: 

```
func AppendExtras(js []byte, id_map []string, paths []string, extras_db *database.SQLiteDatabase) ([]byte, error) {
```

As in:

* A JSON-serialized `spr.StandardPlacesResults` blob of bytes that be queried
* An ordered list of IDs that maps to each item in the `places` list (in the serialized `spr.StandardPlacesResults` blob)
* A list of paths (in dot notation) to look up for each ID (and append to its corresponding `place` record) in a GeoJSON properties dictionary
* A valid `database.SQLiteDatabase` with a `geojson` table following the schema defined by the [go-whosonfirst-sqlite-features](https://github.com/whosonfirst/go-whosonfirst-sqlite-features#geojson) package.

There is also a handy `extras.AppendExtrasWithSPRResults` helper method for
generating the list of IDs required by (and which invokes) the `AppendExtras`
method.

```
func AppendExtrasWithSPRResults(js []byte, results spr.StandardPlacesResults, paths []string, extras_db *database.SQLiteDatabase) ([]byte, error) {
```

For example:

```
	// index := ...
	// coord := ...
	// filter := ...

	results, _ := index.GetIntersectsByCoord(coord, filter)
	js, _ := json.Marshal(results)

	extras_dsn := "extras.db"
	extras_db, _ := database.NewDB(extras_dsn)

	extras_paths := []string{
		"geom:",
	}

	js, _ = extras.AppendExtrasWithSPRResults(js, results, extras_paths, extras_db)
```

A few things to note about "extras":

* Remember: "extras" are still considered experimental. Comments, suggestions
  and gentle cluebats are welcome and encouraged but understand that it's all
  still wet paint.

* This may get replaced by a generic [S3
  Select](https://github.com/whosonfirst/go-whosonfirst-select) -like interface
  which would allow filtering across arbritrary properties. Today that is not
  possible.

* If you are using one of the command line tools and indexing documents using
  `-mode spatialite` then the path for the `-extras-dsn` flag needs to be the same as
  the path for `-spatialite-dsn` flag. You can also just leave the default value
  (`:tmpfile:`) of the `-extras-dsn` flag and the code will update it accordingly.

## Filters

There are 6 different filters, divided in to two classes, for limiting
results. The two classes are: placetypes and existential flags.

There is one placetype flag (called `placetype`) which is defined as any
placetype string.

There are five existential flags: `current`, `deprecated`, `ceased`,
`superseded` and `superseding`. An existential flag can be defined as true or
false (`1` or `0` respectively) or unknown (`-1`).

To filter your query (when calling the `wof-pip-server`) simply pass along one
of more of the following parameters:

* `placetype={PLACETYPE}`
* `is_current={EXISTENTIAL_FLAG}`
* `is_deprecated={EXISTENTIAL_FLAG}`
* `is_ceased={EXISTENTIAL_FLAG}`
* `is_superseded={EXISTENTIAL_FLAG}`
* `is_superseding={EXISTENTIAL_FLAG}`

For example:

```
http://localhost:8080/?latitude=37.6588&longitude=-122.4979&placetype=locality&is_current=1,-1:
```

Under the hood the code is creating a `filter.SPRFilter` thingy (which implements
the `filter.Filter` interface described below) derived from HTTP query
parameters. The details of that process are pretty boring so there is a handy
wrapper method that looks like this:

```
	// req is a *gohttp.Request

	query := req.URL.Query()
	filters, err := filter.NewSPRFilterFromQuery(query)
```

Which produces something that looks like and which this (and is passed to the
`GetIntersectsByCoord` method to limit results):

```
type SPRFilter struct {
	Filter
	Placetypes  []flags.PlacetypeFlag
	Current     []flags.ExistentialFlag
	Deprecated  []flags.ExistentialFlag
	Ceased      []flags.ExistentialFlag
	Superseded  []flags.ExistentialFlag
	Superseding []flags.ExistentialFlag
}
```

## Indexes (indices)

Indexing layers are used to store and query spatial data for performing point in
polygon lookups.

It is important to remember that the indexing layer is populated at (data)
indexing time and only stores the relevant spatial data. By default the index
layer is assumed to be separate and decoupled from the source data, or "input"
layer.

_There is one exception to this rule that is implemented in the `wof-pip*` tools
bundled with this package. If the tools are invoked with the `-mode spatialite`
flag then it will be understood that both the caching and indexing layers
already exist and they will not be pre-populated. This is a piece of
[package-specific helper
code](https://github.com/whosonfirst/go-whosonfirst-pip-v2/blob/spatialite/app/pip.go#L82-L123)
independent of the basic model for creating caches and indices._

### rtree

This is an in-memory RTree implementation that is created during indexing. Under
the hood it uses Daniel Connely's [rtreego package](http://dhconnelly.com/rtreego/) and then
performs a final raycasting operation to filter out false positives.

Indexing time will vary depending on your hardware configuration. In our
experience it is possible to index the entirety of the [Who's On First
administrative data](https://github.com/whosonfirst-data) in about 10-12GB of
RAM, in a little under 10 minutes time.

### spatialite

This is a Spatialite (SQLite with the `libspatialite` extension) based cache that assumes a `geometries` table matching the schema
defined in the
[go-whosonfirst-sqlite-features](https://github.com/whosonfirst/go-whosonfirst-sqlite-features#geometries)
package. It is generally assumed that the databases created by that package will
be used with this caching layer but if you need or want to create your own the
schema looks like this:

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

_This assumes that you have already installed [libspatialite](https://www.gaia-gis.it/fossil/libspatialite/index) on your machine,
the details of which are out of scope for this document._

## Caches

The caching layer is used to persist non-spatial data that needs to be returned
with each result (the SPR) or used to filter queries.

It is important to remember that the caching layer is populated at indexing time
and only stores a feature's `SPR`. By default the caching layer is assumed to be
separate and decoupled from the source data, or "input" layer.

_There is one exception to this rule that is implemented in the `wof-pip*` tools
bundled with this package. If the tools are invoked with the `-mode spatialite`
flag then it will be understood that both the caching and indexing layers
already exist and they will not be pre-populated. This is a piece of
[package-specific helper
code](https://github.com/whosonfirst/go-whosonfirst-pip-v2/blob/spatialite/app/pip.go#L82-L123)
independent of the basic model for creating caches and indices._

### fs

This is a filesystem based cache that stores a feature's `SPR` response in files
on disk, following the [Who's On First URI conventions](https://www.whosonfirst.org/docs/uris/).

### gocache

This is an in-memory cache using Patrick Mylund Nielsen's
[go-cache](https://github.com/patrickmn/go-cache) package that is created during
indexing that stores a feature's `SPR` response (see above).

### spatialite

_This is just an alias of the `sqlite` cache._

### sqlite

This is a SQLite based cache that assumes a `geojson` table matching the schema
defined in the
[go-whosonfirst-sqlite-features](https://github.com/whosonfirst/go-whosonfirst-sqlite-features#geojson)
package. It is generally assumed that the databases created by that package will
be used with this caching layer but if you need or want to create your own the
schema looks like this:

```
CREATE TABLE geojson (
       id INTEGER NOT NULL PRIMARY KEY,
       body TEXT,
       lastmodified INTEGER
);

CREATE INDEX geojson_by_lastmod ON geojson (lastmodified);
```

## Interfaces

This package defines the following interfaces for indexing, caching and filtering layers.

### index.Index

```
type Index interface {
	IndexFeature(geojson.Feature) error
	Cache() cache.Cache
	GetIntersectsByCoord(geom.Coord, filter.Filter) (spr.StandardPlacesResults, error)
	GetCandidatesByCoord(geom.Coord) (*pip.GeoJSONFeatureCollection, error)
	GetIntersectsByPath(geom.Path, filter.Filter) ([]spr.StandardPlacesResults, error)
	Close() error
}
```

`spr.StandardPlacesResult` and `geojson.Feature` are defined as part of the
[go-whosonfirst-spr](https://github.com/whosonfirst/go-whosonfirst-flags) and
[go-whosonfirst-geojson-v2](https://github.com/whosonfirst/go-whosonfirst-geojson-v2)
packages respectively.

### cache.Cache

```
type Cache interface {
	Get(string) (CacheItem, error)
	Set(string, CacheItem) error
	Hits() int64
	Misses() int64
	Evictions() int64
	Size() int64
	Close() error
}
```

### cache.CacheItem

```
type CacheItem interface {
	SPR() spr.StandardPlacesResult
	Polygons() []geojson.Polygon
	Geometry() pip.GeoJSONGeometry
}
```

`spr.StandardPlacesResult` and `geojson.Polygon` are defined as part of the
[go-whosonfirst-spr](https://github.com/whosonfirst/go-whosonfirst-flags) and
[go-whosonfirst-geojson-v2](https://github.com/whosonfirst/go-whosonfirst-geojson-v2)
packages respectively.

`pip.GeoJSONGeometry` is not an interface but rather a local struct defined in
the [pip.go](pip.go) file. A discussion of the many different ways to model
GeoJSON in Go is outside the scope of this document. There are many ways. This
one is ours. It would be awesome if we didn't have to do this...

### cache.FeatureCache

```
type FeatureCache struct {
	CacheItem       `json:",omitempty"`
	FeatureSPR      spr.StandardPlacesResult `json:"spr"`
	FeaturePolygons []geojson.Polygon        `json:"polygons"`
}
```

`spr.StandardPlacesResult` and `geojson.Polygon` are defined as part of the
[go-whosonfirst-spr](https://github.com/whosonfirst/go-whosonfirst-flags) and
[go-whosonfirst-geojson-v2](https://github.com/whosonfirst/go-whosonfirst-geojson-v2)
packages respectively.

### filter.Filter

```
type Filter interface {
	HasPlacetypes(flags.PlacetypeFlag) bool
	IsCurrent(flags.ExistentialFlag) bool
	IsDeprecated(flags.ExistentialFlag) bool
	IsCeased(flags.ExistentialFlag) bool
	IsSuperseded(flags.ExistentialFlag) bool
	IsSuperseding(flags.ExistentialFlag) bool
}
```

`flags.ExistentialFlag` and `flags.PlacetypeFlag` are both defined as part of
the [go-whosonfirst-flags](https://github.com/whosonfirst/go-whosonfirst-flags)
package.

## Example

### Basic

```
package main

import (
       "context"
       "fmt"
       "github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
       "github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/geometry"
       "github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
       wof_index "github.com/whosonfirst/go-whosonfirst-index"
       "github.com/whosonfirst/go-whosonfirst-pip/cache"
       "github.com/whosonfirst/go-whosonfirst-pip/filter"
       "github.com/whosonfirst/go-whosonfirst-pip/index"
       "io"
)

func main() {

	data := "/usr/local/data/whosonfirst-data"
     	mode := "repo"

	gocache_opts, _ := cache.DefaultGoCacheOptions()
	gocache, _ := cache.NewGoCache(gocache_opts)

	rtree_index, _ := index.NewRTreeIndex(gocache)

	cb := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

		f, _ := feature.LoadFeatureFromReader(fh)
		geom_type := geometry.Type(f)

		if geom_type == "Point" {
			return nil
		}

		return rtree_index.IndexFeature(f)
	}

	idx, _ := wof_index.NewIndexer(mode, cb)
	idx.IndexPaths([]{ data })

	// time passes and/or you check the value of idx.IsIndexing()

	c, _ := utils.NewCoordinateFromLatLons(lat, lon)
	f, _ := filter.NewSPRFilter()

	results, _ := rtree_index.GetIntersectsByCoord(c, f)

	body, _ := json.Marshal(results)
	fmt.Println(string(body))
```

_Error handling has been removed for the sake of brevity._

There are a few things to note about the example above:

* See the way the name is still is
  `github.com/whosonfirst/go-whosonfirst-pip/...` even though this package is
  called `github.com/whosonfirst/go-whosonfirst-pip-v2` ? That's unfortunate and
  something that we'll reconcile in the future...

* See the way there is a `github.com/whosonfirst/go-whosonfirst-index` package
  (for indexing data) and a `github.com/whosonfirst/go-whosonfirst-pip/index`
  package (for spatial indexes) ? This is why we can't have nice things...

* See the way we're creating a new `filter.NewSPRFilter()` variable and then
blindly passing it to the `GetIntersectsByCoord` method ? It's possible the
interface for the method will change to take an variable set of arguments but
today it requires a "filter" even if that filter is null.

## Tools

The following tools are included with this package.

#### Command line flags versus environment variables

If you pass the `-setenv` flag to any of the tools below all the flags defined
will be checked for a corresponding environment variable and set
accordingly. Given a flag the rules for mapping it an to environment variables are:

* Upper string the name
* Replace all instances of `-` with `_`
* Prefix the new string with `WOF_`

For example the `-index` and `-spatialite-dsn` flags becomes `WOF_INDEX` and
`WOF_SPATIALITE_DSN` respectively. Like this:

```
$> setenv WOF_INDEX spatialite
$> setenv WOF_CACHE spatialite
$> setenv WOF_SPATIALITE_DSN /usr/local/data/whosonfirst-data-constituency-us-latest.db
$> setenv WOF_ENABLE_WWW true
$. setenv WOF_MODE spatialite

$> ./bin/wof-pip-server -setenv
2018/03/08 10:59:23 set -cache flag (spatialite) from WOF_CACHE environment variable
2018/03/08 10:59:23 set -enable-www flag (true) from WOF_ENABLE_WWW environment variable
2018/03/08 10:59:23 set -index flag (spatialite) from WOF_INDEX environment variable
2018/03/08 10:59:23 set -mode flag (spatialite) from WOF_MODE environment variable
2018/03/08 10:59:23 set -spatialite-dsn flag (/usr/local/data/whosonfirst-data-constituency-us-latest.db) from WOF_SPATIALITE_DSN environment variable
2018/03/08 10:59:23 -enable-www flag is true causing the following flags to also be true: -enable-geojson -enable-candidates
2018/03/08 10:59:23 [WARNING] -enable-www flag is set but -www-api-key is empty
10:59:23.543821 [wof-pip-server] STATUS listening for requests on localhost:8080
```

### wof-pip

`wof-pip` is an interactive tool for querying a set of Who's On First (or GeoJSON) documents.

```
./bin/wof-pip -h
  -cache string
    	Valid options are: gocache, fs, spatialite, sqlite. Note that the spatalite option is just a convenience to mirror the '-index spatialite' option. (default "gocache")
  -cache-all
    	This flag is DEPRECATED and doesn't do anything anymore.
  -exclude value
    	Exclude (WOF) records based on their existential flags. Valid options are: ceased, deprecated, not-current, superseded.
  -failover-cache string
    	This flag is DEPRECATED and doesn't do anything anymore.
  -fs-path string
    	The root directory to look for features if '-cache fs'.
  -index string
    	Valid options are: rtree, spatialite. (default "rtree")
  -is-wof
    	Input data is WOF-flavoured GeoJSON. (Pass a value of '0' or 'false' if you need to index non-WOF documents. (default true)
  -lru-cache-size int
    	This flag is DEPRECATED and doesn't do anything anymore.
  -lru-cache-trigger int
    	This flag is DEPRECATED and doesn't do anything anymore.
  -mode string
    	Valid modes are: directory, feature, feature-collection, files, geojson-ls, meta, path, repo, spatialite, sqlite. (default "files")
  -processes int
    	This flag is DEPRECATED and doesn't do anything anymore.
  -setenv
	Set flags from environment variables.
  -source-cache-root string
    	This flag is DEPRECATED and doesn't do anything anymore. Please use the '-cache fs' and '-fs-path {PATH}' flags instead.
  -spatialite-dsn string
    	A valid SQLite DSN for the '-cache spatialite/sqlite' or '-index spatialite' option. As of this writing for the '-index' and '-cache' options share the same '-spatailite' DSN.
  -strict
    	Be strict about flags and fail if any are missing or deprecated flags are used.
  -verbose
    	Be chatty.
```

For example:

### wof-pip-server

`wof-pip-server` is an HTTP daemon for querying Who's On First (or GeoJSON) documents.

```
./bin/wof-pip-server -h
  -allow-extras
    	This flag is DEPRECATED. Please use the '-enable-extras' flag instead.
  -allow-geojson
    	This flag is DEPRECATED. Please use the '-enable-geojson' flag instead.
  -cache string
    	Valid options are: gocache, fs, spatialite, sqlite. Note that the spatalite option is just a convenience to mirror the '-index spatialite' option. (default "gocache")
  -cache-all
    	This flag is DEPRECATED and doesn't do anything anymore.
  -candidates
    	This flag is DEPRECATED. Please use the '-enable-candidates' flag instead.
  -enable-candidates
    	Enable the /candidates endpoint to return candidate bounding boxes (as GeoJSON) for requests.
  -enable-extras
    	Enable support for 'extras' parameters in queries.
  -enable-geojson
    	Allow users to request GeoJSON FeatureCollection formatted responses.
  -enable-polylines
    	Enable the /polylines endpoint to return hierarchies intersecting a path.
  -enable-www
    	Enable the interactive /debug endpoint to query points and display results.
  -exclude value
    	Exclude (WOF) records based on their existential flags. Valid options are: ceased, deprecated, not-current, superseded.
  -extras-db string
    	This flag is DEPRECATED. Please use '-extras-dsn' flag instead.
  -extras-dsn string
    	A valid SQLite DSN for your 'extras' database - if ':tmpfile:' then a temporary database will be created during indexing and deleted when the program exits. (default ":tmpfile:")
  -failover-cache string
    	This flag is DEPRECATED and doesn't do anything anymore.
  -fs-path string
    	The root directory to look for features if '-cache fs'.
  -host string
    	The hostname to listen for requests on. (default "localhost")
  -index string
    	Valid options are: rtree, spatialite. (default "rtree")
  -is-wof
    	Input data is WOF-flavoured GeoJSON. (Pass a value of '0' or 'false' if you need to index non-WOF documents. (default true)
  -lru-cache-size int
    	This flag is DEPRECATED and doesn't do anything anymore.
  -lru-cache-trigger int
    	This flag is DEPRECATED and doesn't do anything anymore.
  -mapzen-api-key string
    	This flag is DEPRECATED. Please use the '-www-api-key' flag instead.
  -mode string
    	Valid modes are: directory, feature, feature-collection, files, geojson-ls, meta, path, repo, spatialite, sqlite. (default "files")
  -polylines
    	This flag is DEPRECATED. Please use the '-enable-polylines' flag instead.
  -polylines-max-coords int
    	The maximum number of points a (/polylines) path may contain before it is automatically paginated. (default 100)
  -port int
    	The port number to listen for requests on. (default 8080)
  -processes int
    	This flag is DEPRECATED and doesn't do anything anymore.
  -setenv
	Set flags from environment variables.
  -source-cache-root string
    	This flag is DEPRECATED and doesn't do anything anymore. Please use the '-cache fs' and '-fs-path {PATH}' flags instead.
  -spatialite-dsn string
    	A valid SQLite DSN for the '-cache spatialite/sqlite' or '-index spatialite' option. As of this writing for the '-index' and '-cache' options share the same '-spatailite' DSN.
  -strict
    	Be strict about flags and fail if any are missing or deprecated flags are used.
  -verbose
    	Be chatty.
  -www
    	This flag is DEPRECATED. Please use the '-enable-www' flag instead.
  -www-api-key string
    	A valid Nextzen Map Tiles API key (https://developers.nextzen.org). (default "xxxxxx")
  -www-local string
    	This flag is DEPRECATED and doesn't do anything anymore.
  -www-local-root string
    	This flag is DEPRECATED and doesn't do anything anymore.
  -www-path string
    	The URL path for the interactive debug endpoint. (default "/debug")
```

For example, to index [Who's On First data published as a SQLite database](https://dist.whosonfirst.org/sqlite) and spinning up a little web server for debugging things you might do something like:

```
wget https://dist.whosonfirst.org/sqlite/region-20171212.db.bz2
bunzip2 region-20171212.db.bz2
```

And then:

```
./bin/wof-pip-server -index spatialite -cache spatialite -spatialite-dsn region-20171212.db -enable-www -www-api-key **** -mode spatialite
16:37:25.490337 [wof-pip-server] STATUS -enable-www flag is true causing the following flags to also be true: -enable-geojson -enable-candidates
16:37:25.490562 [wof-pip-server] STATUS listening on localhost:8080
16:37:26.491416 [wof-pip-server] STATUS indexing 33 records indexed
16:37:27.495491 [wof-pip-server] STATUS indexing 118 records indexed
16:37:28.490831 [wof-pip-server] STATUS indexing 138 records indexed
16:37:29.490722 [wof-pip-server] STATUS indexing 312 records indexed
...time passes...
16:40:25.496078 [wof-pip-server] STATUS indexing 4691 records indexed
16:40:26.498284 [wof-pip-server] STATUS indexing 4694 records indexed
16:40:27.494674 [wof-pip-server] STATUS indexing 4697 records indexed
16:40:28.494235 [wof-pip-server] STATUS indexing 4900 records indexed
16:40:29.498331 [wof-pip-server] STATUS indexing 4952 records indexed
16:40:29.562617 [wof-pip-server] STATUS finished indexing
```

_Note the part where you need to get a [Nextzen Map Tiles API key](https://developers.nextzen.org/) in order for the map-y parts of things to work._

And finally:

```
open localhost:8080/debug
```

And you should see something like this:

![](docs/images/wof-pip-sqlite.png)

_See the way that screenshot uses the old deprecated flags? We'll fix that soon..._

#### Fancy

Indexing results and then fetching all the places that intersect a polyline:

```
./bin/wof-pip-server -enable-polylines -mode meta /usr/local/data/whosonfirst-data/meta/wof-microhood-latest.csv
11:56:04.605805 [wof-pip-server] STATUS listening on localhost:8080
11:56:05.606812 [wof-pip-server] STATUS indexing 537 records indexed
11:56:06.608378 [wof-pip-server] STATUS indexing 749 records indexed
11:56:07.610900 [wof-pip-server] STATUS indexing 1069 records indexed
11:56:08.609043 [wof-pip-server] STATUS indexing 1298 records indexed
11:56:09.356357 [wof-pip-server][index] STATUS time to index meta file '/usr/local/data/whosonfirst-data/meta/wof-microhood-latest.csv' 4.750478843s
11:56:09.356370 [wof-pip-server][index] STATUS time to index path '/usr/local/data/whosonfirst-data/meta/wof-microhood-latest.csv' 4.750568978s
11:56:09.356374 [wof-pip-server][index] STATUS time to index paths (1) 4.750577455s
11:56:09.356377 [wof-pip-server] STATUS finished indexing
```

_See the way we're indexing a Who's On First `meta` (CSV) file instead of a SQLite database this time?_

And then given a polyline (`oqseF~gcjVvRQaJbLhRuIzN_JeFza@cH{@gK`KxMtErX_NeXtf@yW{l@`) like this:

![](docs/images/wof-pip-polyline.png)

You could do this:

```
curl -s 'localhost:8080/polyline?polyline=oqseF%7EgcjVvRQaJbLhRuIzN_JeFza%40cH%7B%40gK%60KxMtErX_NeXtf%40yW%7Bl%40' | jq '.places[]["wof:name"]'
"The Sit/Lie"
"The Park"
"The Gimlet"
"The Panhandle"
"Financial District South"
"The Post Up"
"Little Saigon"
"Little Saigon"
"The Panhandle"
"Deli Hills"
"The Panhandle"
"Civic Center"
"Van Ness"
```

There are two important things to note here, at least as of this writing:

1. It is left as an exercise to consumers of the `/polyline` endpoint to deduplicate results (assuming you wanted a list of unique places that intersect a polyline)
2. If you are passing in [a polyline line returned from Valhalla's turn-by-turn
routing service](https://github.com/valhalla/valhalla) you will need to include a `?precision=6` query parameter with your request so that the code can properly decode your polyline
2. The response format for the `/polyline` endpoint _will_ change so please don't get too attached to anything that is returned today

See also: https://github.com/whosonfirst/go-mapzen-valhalla#valhalla-route

#### Fancy McFancyPants

_Note: As of this writing the [Who's On First API]() is still offline but the
point – the ability to index line-separated GeoJSON by piping it to the
`wof-pip-server` – remains the same._

Indexing API results (in this case counties in California) by piping them in to `wof-pip-server` on STDIN:

```
../go-whosonfirst-api/bin/wof-api -param api_key=mapzen-xxxxxx \
    -param method=whosonfirst.places.getDescendants -param id=85688637 \
    -param placetype=county -geojson-ls \
    | \
    ./bin/wof-pip-server -mode geojson-ls -enable-www -www-api-key xxxxxx \
    -port 8081 \
    STDIN
    
11:18:19.537724 [wof-pip-server] STATUS listening on localhost:8081
11:18:20.538209 [wof-pip-server] STATUS indexing 0 records indexed
11:18:21.538002 [wof-pip-server] STATUS indexing 2 records indexed
11:18:22.538104 [wof-pip-server] STATUS indexing 4 records indexed
...
11:18:45.537952 [wof-pip-server] STATUS indexing 51 records indexed
11:18:46.538419 [wof-pip-server] STATUS indexing 54 records indexed
11:18:47.539162 [wof-pip-server] STATUS indexing 57 records indexed
11:18:47.736253 [wof-pip-server][index] STATUS time to index geojson-ls 'STDIN' 28.198454417s
11:18:47.736274 [wof-pip-server][index] STATUS time to index path 'STDIN' 28.198542171s
11:18:47.736282 [wof-pip-server][index] STATUS time to index paths (1) 28.198563784s
11:18:47.736288 [wof-pip-server] STATUS finished indexing
```

![](docs/images/wof-pip-counties.png)

## Plain old GeoJSON

Let assume that you've downloaded the [OSM water polygons data](http://openstreetmapdata.com/data/water-polygons) and created a GeoJSON file. For example:

```
cd /usr/local
wget http://data.openstreetmapdata.com/water-polygons-split-4326.zip
unzip water-polygons-split-4326.zip
cd water-polygons-split-4326
ogr2ogr -F GeoJSON water_polygons.geojson water_polygons.shp
```

Now we start up the PIP server passing along the `-is-wof=false` and `-mode feature-collection` flag:

```
./bin/wof-pip-server -port 5555 -is-wof=false -enable-www -www-api-key xxxxxxx \
	-mode feature-collection /usr/local/water-polygons-split-4326/water_polygons.geojson

10:33:49.735255 [wof-pip-server] STATUS -www flag is true causing the following flags to also be true: -allow-geojson -candidates
10:33:49.735402 [wof-pip-server] STATUS listening on localhost:5555
10:33:50.735473 [wof-pip-server] STATUS indexing 0 records indexed
10:33:51.736308 [wof-pip-server] STATUS indexing 0 records indexed
10:33:52.735489 [wof-pip-server] STATUS indexing 0 records indexed
...
10:47:52.756351 [wof-pip-server] STATUS indexing 41239 records indexed
10:47:53.755285 [wof-pip-server] STATUS indexing 41311 records indexed
10:47:54.754969 [wof-pip-server] STATUS indexing 41376 records indexed
10:47:55.757847 [wof-pip-server] STATUS indexing 41405 records indexed
10:47:56.479703 [wof-pip-server][index] STATUS time to index feature collection '/usr/local/water-polygons-split-4326/water_polygons.geojson' 14m6.724468767s
10:47:56.479721 [wof-pip-server][index] STATUS time to index path '/usr/local/water-polygons-split-4326/water_polygons.geojson' 14m6.725233324s
10:47:56.479725 [wof-pip-server][index] STATUS time to index paths (1) 14m6.725246195s
10:47:56.479729 [wof-pip-server] STATUS finished indexing
```

And then you would query it as usual:

```
curl 'http://localhost:5555/?latitude=54.793624&longitude=-79.948933&format=geojson' \
{"type":"FeatureCollection","features":[{"type":"Feature","geometry":{"type":"MultiPolygon","coordinates":[[[[-80.15635,54.140525],[-80.15635,54.84385],[-79.8045875,54.84385],[-79.8045875,54.72711395185386],[-79.8061901,54.7266501],[-79.8070701,54.7267899],[-79.8075399,54.7257701],[-79.8102099,54.72528],[-79.81023,54.7247701],[-79.8111099,54.7249099],[-79.8119999,54.72453],[-79.81206,54.72274],[-79.8045875,54.724551416699605],[-79.8045875,54.140525],[-80.15635,54.140525]],[],[],[[-79.93955,54.4322999],[-79.93886,54.4331899],[-79.9390599,54.43422],[-79.943,54.4345099],[-79.94302,54.4337399],[-79.94258,54.4337299],[-79.9417399,54.43257],[-79.93955,54.4322999]],[[-79.83404,54.7429299],[-79.82869,54.74429],[-79.8282299,54.7449199],[-79.82735,54.7449199],[-79.8273199,54.74594],[-79.8268701,54.74594],[-79.82729,54.74696],[-79.82818,54.74672],[-79.8281901,54.7462],[-79.8299701,54.74597],[-79.8304399,54.7452],[-79.8317699,54.74521],[-79.8326799,54.7442],[-79.8344499,54.7442201],[-79.8341901,54.74328],[-79.83449,54.74295],[-79.83404,54.7429299]]]]},"properties":{"spr:id":"f1xnh0000000","spr:name":"f1xnh0000000","spr:placetype":"polygon","spr:latitude":54.4921875,"spr:longitude":-79.98046875,"spr:min_latitude":54.140525,"spr:min_longitude":-80.15635,"spr:max_latitude":54.84385,"spr:max_longitude":-79.8045875}}],"pagination":{"total_count":0,"page":0,"per_page":0,"page_count":0}}
```

![](docs/images/wof-pip-water-polygons.png)

### Fancy

If you want to get fancy about things you could also do something like this:

```
./bin/wof-pip-server -index spatialite -cache spatialite -spatialite-dsn water.db -port 5555 -is-wof=false -enable-www -www-api-key xxxxxx \
   -mode feature-collection water-polygons-split-4326/water_polygons.geojson

...time passes...

12:08:12.171701 [wof-pip-server] STATUS finished indexing in 20m4.457747712s
```

Which tells `wof-pip-server` to use a Spatialite index and cache (defined by
`-spatialite-dsn water.db` flag) to store all the data indexed from GeoJSON
files (as defined by the `-mode feature-collection` and
`water-polygons-split-4326/water_polygons.geojson` flags).

That means that the next time you want to start `wof-pip-server` you can _skip
the indexing stage_ and simply do this instead:

```
./bin/wof-pip-server -index spatialite -cache spatialite -spatialite-dsn water.db \
   -port 5555 -enable-www -www-api-key xxxxxx -mode spatialite`

...time passes faster (because there is no indexing phase)...
```

### Caveats

If you're using plain-old GeoJSON you should expect filters (for both placetypes and existential fields) to be weird. For example:

```
2018/03/09 12:09:17 Unable to parse placetype (multipolygon) for ID 5590284368062970320, because 'Invalid placetype' - skipping placetype filters
2018/03/09 12:09:19 Unable to parse placetype (multipolygon) for ID 5589424534403940352, because 'Invalid placetype' - skipping placetype filters
2018/03/09 12:09:19 Unable to parse placetype (multipolygon) for ID 5590284368062970320, because 'Invalid placetype' - skipping placetype filters
```

We should make this "less weird" going forward but today it is "weird".

## Docker

[Yes](Dockerfile), although it's still early days and should still be considered
experimental. The biggest open question is how to manage data files which can be
very large.

Currently the `Dockerfile` itself simply sets up dependencies and creates a
volume named `/usr/local/data` and then hands everything off to a
[docker/entrypoint.sh](docker/entrypoint.sh) shell script that looks for things
to download and index from [dist.whosonfirst.org](https://dist.whosonfirst.org)
_when the Docker instance is started_.

Currently the approach only works with the Who's On First SQLite
databases. Bundles and other remote download sources are _not supported_
yet.

To build the Docker image you do the usual `docker build` dance like this:

```
docker build -t wof-pip-server .
```

To start the image you do the usual `docker run` dance passing one or more
`WOF_` environment variables ([as described above](#command-line-flags-versus-environment-variables)).

### sqlite

If your `WOF_MODE` environment variable is `sqlite` then you need to also set a
`SQLITE_DATABASES` environment variable containing a comma-separated list of
(WOF) SQLite database names (including the trailing `.db`) to fetch and index.

```
> docker run -it -p 6161:8080 -e WOF_HOST='0.0.0.0' -e WOF_ENABLE_EXTRAS='true' -e WOF_MODE='sqlite' -e SQLITE_DATABASES='whosonfirst-data-constituency-ca-latest.db' wof-pip-server
fetch https://dist.whosonfirst.org/sqlite/whosonfirst-data-constituency-ca-latest.db.bz2
2018/03/09 15:57:56 set -enable-extras flag (true) from WOF_ENABLE_EXTRAS environment variable
2018/03/09 15:57:56 set -host flag (0.0.0.0) from WOF_HOST environment variable
2018/03/09 15:57:56 set -mode flag (sqlite) from WOF_MODE environment variable
15:57:56.155541 [wof-pip-server] STATUS listening for requests on 0.0.0.0:8080
15:57:56.863973 [wof-pip-server] STATUS finished indexing in 708.871526ms
15:57:57.317221 [wof-pip-server] STATUS indexing 42 records indexed
15:57:58.350197 [wof-pip-server] STATUS indexing 47 records indexed
15:57:59.274145 [wof-pip-server] STATUS indexing 52 records indexed
15:58:00.204993 [wof-pip-server] STATUS indexing 60 records indexed

{
  "places": [
    {
      "geom:area": 0.003125,
      "geom:area_square_m": 25188051.658008,
      "geom:bbox": "-123.148024893,49.2946242012,-123.019433551,49.3357115812",
      "geom:latitude": 49.314573,
      "geom:longitude": -123.077469,
      "mz:is_ceased": -1,
      "mz:is_current": -1,
      "mz:is_deprecated": 0,
      "mz:is_superseded": 0,
      "mz:is_superseding": 0,
      "mz:latitude": 49.314573,
      "mz:longitude": -123.077469,
      "mz:max_latitude": 49.33571158121562,
      "mz:max_longitude": -123.01943355128094,
      "mz:min_latitude": 49.29462420121845,
      "mz:min_longitude": -123.1480248934407,
      "mz:uri": "https://data.whosonfirst.org/110/896/285/1/1108962851.geojson",
      "wof:country": "CA",
      "wof:id": 1108962851,
      "wof:lastmodified": 1494447496,
      "wof:name": "North Vancouver-Lonsdale",
      "wof:parent_id": -1,
      "wof:path": "110/896/285/1/1108962851.geojson",
      "wof:placetype": "constituency",
      "wof:repo": "whosonfirst-data-constituency-ca",
      "wof:superseded_by": [],
      "wof:supersedes": []
    }
  ]
}
```

### spatialite

_IMPORTANT - THIS DOESN'T ACTUALLY WORK YET AND I AM INCLUDING IT HERE IN THE HOPES THAT SOMEONE CAN TELL ME WHAT I AM DOING WRONG._

Specifically in the [Dockerfile](Dockerfile) we are using Alpine Linux and trying to install the spatialiate library like this:

```
RUN apk add --update --repository http://dl-3.alpinelinux.org/alpine/edge/testing/ libspatialite
```

But when we actually try to load it (in our Go code) it fails with the following error which is invoked [over here](https://github.com/whosonfirst/go-whosonfirst-sqlite/blob/master/vendor/github.com/whosonfirst/go-spatialite/spatialite.go):

```
Failed to create new PIP application, because shaxbee/go-spatialite: spatialite extension not found.
```

It is unclear to me what the problem is, but it should work like this (modulo the errors):

---

If your `WOF_MODE` environment variable is `spatialite` then you need to also
set a `WOF_SPATIALITE_DATABASE` containing the name of a (WOF) SQLite database
name (including the trailing `.db`) to fetch and index.

```
> docker run -it -p 6161:8080 -e WOF_HOST='0.0.0.0' -e WOF_INDEX='spatialite' -e WOF_CACHE='spatialite' -e WOF_MODE='spatialite' -e SPATIALITE_DATABASE='whosonfirst-data-constituency-us-latest.db' wof-pip-server
fetch https://dist.whosonfirst.org/sqlite/whosonfirst-data-constituency-us-latest.db.bz2 as /usr/local/data/whosonfirst-data-constituency-us-latest.db.bz2
2018/03/09 15:50:53 set -cache flag (spatialite) from WOF_CACHE environment
variable
2018/03/09 15:50:53 set -host flag (0.0.0.0) from WOF_HOST environment variable
2018/03/09 15:50:53 set -index flag (spatialite) from WOF_INDEX environment variable
2018/03/09 15:50:53 set -mode flag (spatialite) from WOF_MODE environment variable
2018/03/09 15:50:53 set -spatialite-dsn flag (/usr/local/data/whosonfirst-data-constituency-us-latest.db) from WOF_SPATIALITE_DSN environment variable 
2018/03/09 15:50:53 Failed to create new PIP application, because shaxbee/go-spatialite: spatialite extension not found.
command '/bin/wof-pip-server ' failed
```

### Caveats

Consider this approach a generic proof-of-concept. If you know what data you're
going to be working with ahead of time it probably makes more sense for you to
clone the existing `Dockerfile` and change this:

```
VOLUME /usr/local/data
```

To something like this:

```
ADD your-data /usr/local/data
```

And then update the `ENTRYPOINT` command accordingly to point to the relevant data.

## Performance

Proper performance and load-testing figures still need to be compiled but this is what happened when I ran `siege` with 200 concurrent clients reading from the [testdata/urls.txt](testdata) file and then forgot about it for a week:

```
$> siege -v -c 200 -i -f testdata/urls.txt
...time passes...
Transactions:              267175219 hits
Availability:                 100.00 %
Elapsed time:              686483.75 secs
Data transferred:         1200409.12 MB
Response time:                  0.01 secs
Transaction rate:             389.19 trans/sec
Throughput:                     1.75 MB/sec
Concurrency:                    5.32
Successful transactions:   267175219
Failed transactions:               0
Longest transaction:            0.97
Shortest transaction:           0.00
```

## See also

* https://github.com/whosonfirst/go-whosonfirst-geojson-v2
* https://github.com/whosonfirst/go-whosonfirst-spr
* https://github.com/whosonfirst/go-whosonfirst-flags
* https://github.com/whosonfirst/go-whosonfirst-index
* https://github.com/whosonfirst/go-whosonfirst-pip
* https://github.com/whosonfirst/go-whosonfirst-placetypes
