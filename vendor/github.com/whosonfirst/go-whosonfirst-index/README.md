# go-whosonfirst-index

Go package for indexing Who's On First documents

## Install

You will need to have both `Go` (specifically a version [1.12](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make tools
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Example

```
package main

import (
       "context"
       "flag"
       "github.com/whosonfirst/go-whosonfirst-index"
       _ "github.com/whosonfirst/go-whosonfirst-index/fs"              
       "io"
       "log"
)

func main() {

	var dsn = flag.String("dsn", "repo://", "A valid go-whosonfirst-index DSN")
	
     	flag.Parse()
	
	cb := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

		path, _ := index.PathForContext(ctx)

		log.Println("PATH", path)
		return nil
	}

	i, _ := index.NewIndexer(*dsn, cb)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	paths := flag.Args()

	i.Index(ctx, paths...)
}	
```

_Error handling removed for the sake of brevity._