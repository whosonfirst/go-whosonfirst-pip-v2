# go-mapzen-js

Go middleware package for mapzen.js
 
## Important

This package is no longer being maintained. You should use [go-http-nextzenjs](https://github.com/whosonfirst/go-http-nextzenjs) instead.

## Install

You will need to have both `Go` (specifically a version of Go more recent than 1.6 so let's just assume you need [Go 1.8](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make bin
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## A word about "Mapzen" and naming things

[Mapzen is no more](https://mapzen.com/blog/shutdown/). As I write this I am not really sure what the state of
[nextzen.js](https://github.com/nextzen/nextzen.js) (formerly `mapzen.js`)
is. To top it off all the map vector tiles are now called _Tilezen_ but are
hosted under the [Nextzen](https://www.nextzen.org/) domain. The same is true of
[tangram.js](https://github.com/tangrams/tangram).
 
It's a bit confusing but so is life.

So while this package bundles and exposes a copy of the old `mapzen.js` it
_won't work_ as you'd normally expect, like this:

```

	// remember this data attribute is squirted in to the source via
	// the MapzenJSHandler (.go) handler

	var body = document.body;
	var api_key = body.getAttribute("data-mapzen-api-key");

	L.Mapzen.apiKey = api_key;
			
	var map_opts = { tangramOptions: {
		scene: L.Mapzen.BasemapStyles.Refill
	}};
			
	map = L.Mapzen.map('map', map_opts);
```

Or at least I'm _not sure_ it will, as I write this. Instead what I (currently)
do is this:

```

	// remember this data attribute is squirted in to the source via
	// the MapzenJSHandler (.go) handler

	var body = document.body;
	var api_key = body.getAttribute("data-mapzen-api-key");

	var sources = {
	    mapzen: {
		url: 'https://{s}.tile.nextzen.org/tilezen/vector/v1/512/all/{z}/{x}/{y}.topojson',
		url_subdomains: ['a', 'b', 'c', 'd'],
		url_params: {
		    api_key: api_key	// not clear this actually works... ?
		},
		tile_size: 512,
		max_zoom: 15
	    }
	};
			
	var scene = {
	    import: [
		     "/tangram/refill-style.zip",
		     "/tangram/refill-style-themes-label.zip",
		     ],
	    sources: sources,
	    global: {
		sdk_mapzen_api_key: api_key,
	    },
	};
			
	var attributions = {
	    "Tangram": "https://github.com/tangrams/",
	    "© OSM contributors": "http://www.openstreetmap.org/",
	    "Who\"s On First": "http://www.whosonfirst.org/",
	    "Nextzen": "https://nextzen.org/",
	};
			
	var attrs = [];
			
	for (var label in attributions){
			    
	    var link = attrs[label];
			    
	    if (! link){
		attrs.push(label);
		continue;
	    }
			    
	    var anchor = '<a href="' + link + '" target="_blank">' + enc_label + '</a>';
	    attrs.push(anchor);
	}
			
	var str_attributions = attrs.join(" | ");
			
	// waiting for nextzen.js to be updated to do all the things - that said it's
	// not entirely clear we need all of (map/next)zen.js and could probably get
	// away with leaflet + tangram but for now we'll just keep on as-is...
	// (20180304/thisisaaronland)
			
	L.Mapzen.apiKey = api_key;
			
	var map_opts = {
	    tangramOptions: {
		scene: scene,
		attribution: attributions,
	    }
	};
	
	map = L.Mapzen.map('map', map_opts);
```			

Which, you know, arguably could be wrapped in a helper function/library exposed
by the `MapzenJSAssetsHandler` handler but presumably `nextzen.js` will (has
been) updated to point to all the new Nextzen things, right?

To further confuse matters it's on my list to update this package to bundle and
expose plain vanilla [Leaflet](https://leafletjs.com) support for Tangram and
Tilezen stuff but that hasn't happened yet.

So maybe this package should be called `go-http-nextzen`... but it
isn't... yet... probably.

We'll figure it out soon but honestly, it's kind of a wonder _anyone_ figured
out how to use our (Mapzen) stuff at all. Which is why everything in this
package is still called "mapzen".

## Handlers

### MapzenJSHandler(http.Handler, MapzenJSOptions) (http.Handler, error)

This handler will optionally modify the output of the `your_handler http.Handler` as follows:

* Append the relevant [mapzen.js](https://mapzen.com/documentation/mapzen-js/) `script` and `link` elements to the `head` element.
* Append a `data-mapzen-api-key` attribute (and value) to the `body` element.

```
import (
	"github.com/whosonfirst/go-http-mapzenjs"
	"net/http"
)

func main(){

	opts := mapzenjs.DefaultMapzenJSOptions()
	opts.APIKey = "mapzen-1a2b3c"

	www_handler := YourDefaultWWWHandler()
	
	mapzenjs_handler, _ := mapzenjs.MapzenJSHandler(www_handler, opts)

	mux := http.NewServeMux()
	mux.Handle("/", mapzenjs_handler)
```

_Note that error handling has been removed for the sake of brevity._

#### MapzenJSOptions

The definition for `MapzenJSOptions` looks like this:

```
type MapzenJSOptions struct {
	AppendAPIKey bool
	AppendJS     bool
	AppendCSS    bool
	APIKey       string
	JS           []string
	CSS          []string
}
```

Default `MapzenJSOptions` are:

```
	opts := MapzenJSOptions{
		AppendAPIKey: true,
		AppendJS:     true,
		AppendCSS:    true,
		APIKey:       "mapzen-xxxxxx",
		JS:           []string{"/javascript/mapzen.min.js"},
		CSS:          []string{"/css/mapzen.js.css"},
	}
```

#### Example

Given the following markup generated by your `http.Handler` output:

```
<!DOCTYPE html>
<html lang="en">
  <head>
	  <title>Example</title>
	  <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
	  <meta name="referrer" content="origin">
	  <meta http-equiv="X-UA-Compatible" content="IE=9" />
	  <meta name="apple-mobile-web-app-capable" content="yes" />
	  <meta name="apple-mobile-web-app-status-bar-style" content="black" />
	  <meta name="HandheldFriendly" content="true" />
	  <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, minimum-scale=1, user-scalable=no" />
  </head>
  <body>
  <!-- and so on... ->
```

The `MapzenJSHandler` handler will modify that markup to return:

```
<!DOCTYPE html>
<html lang="en">
  <head>
	  <title>Example</title>
	  <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
	  <meta name="referrer" content="origin">
	  <meta http-equiv="X-UA-Compatible" content="IE=9" />
	  <meta name="apple-mobile-web-app-capable" content="yes" />
	  <meta name="apple-mobile-web-app-status-bar-style" content="black" />
	  <meta name="HandheldFriendly" content="true" />
	  <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, minimum-scale=1, user-scalable=no" />
	  <script type="text/javascript" src="/javascript/mapzen.min.js"></script>
	  <link rel="stylesheet" type="text/css" href="/css/mapzen.js.css" />
  </head>
  <body data-mapzen-api-key="mapzen-1a2b3c">
  <!-- and so on... ->
```

### MapzenJSAssetsHandler() (http.Handler, error)

The handler will serve [mapzen.js](https://mapzen.com/documentation/mapzen-js/) and [tangram.js](https://github.com/tangrams/tangram) related assets which have been bundled with this package.

```
import (
	"github.com/whosonfirst/go-http-mapzenjs"
	"net/http"
)

func main(){

	mapzenjs_assets_handler, _ := mapzen.MapzenJSAssetsHandler()

	mux := http.NewServeMux()

	mux.Handle("/javascript/mapzen.js", mapzenjs_handler)
	mux.Handle("/javascript/mapzen.min.js", mapzenjs_handler)
	mux.Handle("/javascript/tangram.js", mapzenjs_handler)	
	mux.Handle("/javascript/tangram.min.js", mapzenjs_handler)
	mux.Handle("/css/mapzen.js.css", mapzenjs_handler)
	mux.Handle("/tangram/refill-style.zip", mapzenjs_handler)

}
```

You can update the various `mapzen.js` and `tangram.js` assets manually by invoking the `build` target in the included [Makefile](Makefile).

#### Styles

Currently the following styles are bundled with this package:

* [refill](https://tangrams.github.io/refill-style/)
* [walkabout](https://tangrams.github.io/walkabout-style/)

## To do

* Add a tile caching proxy

## See also 

* https://mapzen.com/documentation/mapzen-js/
* https://github.com/tangrams/tangram
