# go-whosonfirst-index

Go package for indexing Who's On First documents

## Install

You will need to have both `Go` (specifically a version of Go more recent than 1.7 so let's just assume you need [Go 1.9](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make bin
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Example

```
package main

import (
       "context"
       "flag"
       "github.com/whosonfirst/go-whosonfirst-index"       
       "io"
       "log"
)

func main() {

	var mode = flag.String("mode", "repo", "A valid go-whosonfirst-index mode")
	
     	flag.Parse()
	
	f := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

		path, err := index.PathForContext(ctx)

		if err != nil {
			return err
		}

		log.Println("PATH", path)
		return nil
	}

	i, err := index.NewIndexer(*mode, f)

	if err != nil {
		log.Fatal(err)
	}

	for _, path := range flag.Args() {

		err := i.IndexPath(path)

		if err != nil {
			log.Fatal(err)
		}
	}
}	
```

## Modes

### directory

Index all the files in a directory.

### feature

Index a GeoJSON Feature. 

### feature-collection

Index all the features in GeoJSON FeatureCollection.

### files

Index a list of files.

### geojson-ls

Index all the features in line-separated GeoJSON list.

### meta

Index all the files listed in a Who's On First "meta" (CSV) file.

### path

Index a path.

### repo

Index all the files in the `data` directory of a Who's On First repository.

### sqlite

Index all the records in the `geojson` table of a Who's On First SQLite database.

