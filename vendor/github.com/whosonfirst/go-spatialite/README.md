# go-spatialite

Go package to enable `libspatialite` support with the [go-sqlite3](https://github.com/mattn/go-sqlite3) `database/sql` driver.

## Important

This is known not to work right now. Specifically, fatal errors [are triggered](https://github.com/whosonfirst/go-whosonfirst-spatialite-geojson/issues/3) on initialization. It's not clear what the problem is or where. Like is it libspatialite module/shared library thing post v4.2 or it is a SQLite extension loading thing or... ?

## See also

* https://sqlite.org/loadext.html
* https://www.gaia-gis.it/fossil/libspatialite/wiki?name=mod_spatialite
* https://github.com/mattn/go-sqlite3#extensions