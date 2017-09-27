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

Indexing results and then fetching all the places that intersect a polyline:

```
./bin/wof-pip-server -polylines -mode meta /usr/local/data/whosonfirst-data/meta/wof-microhood-latest.csv
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
2. If you are passing in [a polyline line returned from Mapzen's Turn-By-Turn service](https://mapzen.com/documentation/mobility/decoding/) you will need to include a `?valhalla=1` query parameter with your request so that the code can properly decode your polyline
2. The response format for the `/polyline` endpoint _will_ change so please don't get too attached to anything that is returned today

#### Fancy McFancyPants

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
