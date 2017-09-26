# go-whosonfirst-pip-v2

An in-memory point-in-polygon (reverse geocoding) package for GeoJSON data, principally Who's On First data.

_This package supersedes the [go-whosonfirst-pip](https://github.com/whosonfirst/go-whosonfirst-pip) package which is (almost, if you're reading this) no longer maintained._

## Install

You will need to have both `Go` (specifically a version of Go more recent than 1.7 so let's just assume you need [Go 1.9](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make bin
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Example

_Please write me._

## Tools

_Please write me._

### wof-pip-server

#### Fancy

Indexing API results (in this case counties in California) by piping them in to `wof-pip-server` on STDIN:

```
../go-whosonfirst-api/bin/wof-api -param api_key=mapzen-xxxxxx -param method=whosonfirst.places.getDescendants -param id=85688637 -param placetype=county -geojson-ls | ./bin/wof-pip-server -mode geojson-ls -www -mapzen-api-key mapzen-xxxxxx -cache gocache -port 8081 STDIN
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

## See also

* https://github.com/whosonfirst/go-whosonfirst-pip
