# go-mapzen-whosonfirst-crawl

Go tools and libraries for crawling a directory of Who's On First data

## Install

You will need to have both `Go` (specifically version [1.12](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make tools
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Example

### crawl.Crawl

```
package main

import (
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-crawl"
	"log"
	"os"
	"time"
	"sync/atomic"
)

func main() {

	root := flag.String("root", "", "The root directory you want to crawl")

	flag.Parse()

	var files int64
	var dirs int64
	
	callback := func(path string, info os.FileInfo) error {
		
		if info.IsDir() {
			atomic.AddInt64(&dirs, 1)
			return nil
		}
		
		atomic.AddInt64(&files, 1)			
		return nil
	}
	
	t0 := time.Now()
	
	defer func(){
		t1 := float64(time.Since(t0)) / 1e9
		fmt.Printf("walked %d files (and %d dirs) in %s in %.3f seconds\n", files, dirs, *root, t1)
	}()
	
	c := crawl.NewCrawler(*root)
	err := c.Crawl(callback)

	if err != nil {
		log.Fatal(err)
	}
}
```

### crawl.CrawlWithContext

_Please write me_

### crawl.CrawlWithChannels

_Please write me_

## Tools

### wof-count

```
./bin/wof-count /usr/local/data/sfomuseum-data-flights-2019-*
go build -o bin/wof-count cmd/wof-count/main.go
count files and directories in  /usr/local/data/sfomuseum-data-flights-2019-01
walked 98116 files (and 0 dirs) in 4.203 seconds
count files and directories in  /usr/local/data/sfomuseum-data-flights-2019-02
walked 87989 files (and 0 dirs) in 4.969 seconds
count files and directories in  /usr/local/data/sfomuseum-data-flights-2019-03
walked 102354 files (and 0 dirs) in 6.178 seconds
count files and directories in  /usr/local/data/sfomuseum-data-flights-2019-04
walked 107775 files (and 0 dirs) in 6.377 seconds
count files and directories in  /usr/local/data/sfomuseum-data-flights-2019-05
walked 143200 files (and 0 dirs) in 7.565 seconds
count files and directories in  /usr/local/data/sfomuseum-data-flights-2019-06
walked 124490 files (and 0 dirs) in 6.326 seconds
count files and directories in  /usr/local/data/sfomuseum-data-flights-2019-07
walked 63 files (and 0 dirs) in 0.007 seconds
count files and directories in  /usr/local/data/sfomuseum-data-flights-2019-08
walked 43 files (and 0 dirs) in 0.005 seconds
```

## See also

* https://github.com/whosonfirst/walk