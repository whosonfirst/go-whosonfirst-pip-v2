// package uri provides methods for parsing and constructing URIs for Who's On First documents.
//
// Example (simple)
//
//	import (
//		"github.com/whosonfirst/go-whosonfirst-uri"
//	)
//
//	fname, _ := uri.Id2Fname(101736545)
//	rel_path, _ := uri.Id2RelPath(101736545)
//	abs_path, _ := uri.Id2AbsPath("/usr/local/data", 101736545)
//
// Produces:
//
//	101736545.geojson
//	101/736/545/101736545.geojson
//	/usr/local/data/101/736/545/101736545.geojson
//
// Example (fancy)
//
//	import (
//		"github.com/whosonfirst/go-whosonfirst-uri"
//	)
//
//	source := "mapzen"
//	function := "display"
//	extras := []string{ "1024" }
//
//	args := uri.NewAlternateURIArgs(source, function, extras...)
//
//	fname, _ := uri.Id2Fname(101736545, args)
//	rel_path, _ := uri.Id2RelPath(101736545, args)
//	abs_path, _ := uri.Id2AbsPath("/usr/local/data", 101736545, args)
//
//Produces:
//
//	101736545-alt-mapzen-display-1024.geojson
//	101/736/545/101736545-alt-mapzen-display-1024.geojson
//	/usr/local/data/101/736/545/101736545-alt-mapzen-display-1024.geojson
//
// For detailed description of the rules governing alternate geometries please consult:
// https://github.com/whosonfirst/whosonfirst-cookbook/blob/master/how_to/creating_alt_geometries.md
package uri
