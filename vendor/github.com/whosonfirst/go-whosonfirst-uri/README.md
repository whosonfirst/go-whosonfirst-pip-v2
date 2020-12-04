# go-whosonfirst-uri

Go package for working with URIs for Who's On First documents

## Install

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Example

### Simple

```
import (
	"github.com/whosonfirst/go-whosonfirst-uri"
)

fname, _ := uri.Id2Fname(101736545)
rel_path, _ := uri.Id2RelPath(101736545)
abs_path, _ := uri.Id2AbsPath("/usr/local/data", 101736545)
```

Produces:

```
101736545.geojson
101/736/545/101736545.geojson
/usr/local/data/101/736/545/101736545.geojson
```

### Fancy

```
import (
	"github.com/whosonfirst/go-whosonfirst-uri"
)

source := "mapzen"
function := "display"
extras := []string{ "1024" }

args := uri.NewAlternateURIArgs(source, function, extras...)

fname, _ := uri.Id2Fname(101736545, args)
rel_path, _ := uri.Id2RelPath(101736545, args)
abs_path, _ := uri.Id2AbsPath("/usr/local/data", 101736545, args)
```

Produces:

```
101736545-alt-mapzen-display-1024.geojson
101/736/545/101736545-alt-mapzen-display-1024.geojson
/usr/local/data/101/736/545/101736545-alt-mapzen-display-1024.geojson
```

## The Long Version

Please read this: https://github.com/whosonfirst/whosonfirst-cookbook/blob/master/how_to/creating_alt_geometries.md

## Tools

### wof-uri-expand

Expand one or more IDs in to their URIs (relative or absolute).

```
./bin/wof-uri-expand -h
Usage of ./bin/wof-uri-expand:
  -root string
    	An optional (filesystem) root to prepend URIs with
  -stdin
    	Read IDs from STDIN
```

For example:

```
./bin/wof-uri-expand 1234556 46632
123/455/6/1234556.geojson
466/32/46632.geojson
```

Or:

```
./bin/wof-uri-expand -root /usr/local/data 1234556 46632
/usr/local/data/123/455/6/1234556.geojson
/usr/local/data/466/32/46632.geojson
```

Or:

```
cat ./test | ./bin/wof-uri-expand -stdin
487/463/636/453/487463636453.geojson
982/635/344/2/9826353442.geojson
```

## See also

* https://github.com/whosonfirst/whosonfirst-cookbook/blob/master/how_to/creating_alt_geometries.md
* https://github.com/whosonfirst/py-mapzen-whosonfirst-uri
