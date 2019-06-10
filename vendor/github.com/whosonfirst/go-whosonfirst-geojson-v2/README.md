# go-whosonfirst-geojson-v2

Go tools for working with Who's On First documents

## Install

You will need to have both `Go` (specifically [version 1.12](https://golang.org/dl/) or higher because we're using [Go modules](https://github.com/golang/go/wiki/Modules)) and the `make` programs installed on your computer. Assuming you do just type:

```
make tools
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Important

This is work in progress. It may change (and break your code) still. This package aims to replace the existing [go-whosonfirst-geojson](https://github.com/whosonfirst/go-whosonfirst-geojson) package. If you want to follow along, please consult:

https://github.com/whosonfirst/go-whosonfirst-geojson-v2/issues/1

## geojson-v2?

Yeah, I don't really like it either but this package is basically 100% backwards incompatible with `github.com/whosonfirst/go-whosonfirst-geojson` and while I don't _really_ think anyone else is using it I don't like the idea of suddenly breaking everyone's code.

## Interfaces

Unlike the first `go-whosonfirst-geojson` package this one at least attempts to define a simplified interface for working with GeoJSON features. These are still in flux.

_Please finish writing me._

### geojson.Feature

```
type Feature interface {
	Type() string
	Id() int64
	Name() string
	Placetype() string
	ToString() string
	ToBytes() []byte
	BoundingBoxes() (BoundingBoxes, error)
	Polygons() ([]Polygon, error)
	ContainsCoord(geom.Coord) (bool, error)
}
```

### geojson.BoundingBoxes

```
type BoundingBoxes interface {
	Bounds() []*geom.Rect
	MBR() geom.Rect
}
```

### geojson.Centroid

```
type Centroid interface {
	Coord() geom.Coord
	Source() string
}
```

### geojson.Polygon

```
type Polygon interface {
	ExteriorRing() geom.Polygon
	InteriorRings() []geom.Polygon
	ContainsCoord(geom.Coord) bool
}
```

## Usage

### Simple

```
import (
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/whosonfirst"
	"log"
)

func main() {

	path := "/usr/local/data/whosonfirst-data/data/101/736/545/101736545.geojson"
	f, err := whosonfirst.LoadFeatureFromFile(path)

	if err != nil {
		log.Fatal(err)
	}

	// prints "Montreal"
	log.Println("Name is ", f.Name())
}
```

## See also

* github.com/skelterjohn/geom
* https://github.com/whosonfirst/go-whosonfirst-geojson

