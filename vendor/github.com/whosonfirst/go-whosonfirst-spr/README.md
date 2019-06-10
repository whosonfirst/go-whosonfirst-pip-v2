# go-whosonfirst-spr

Go tools for working Who's On First "standard places responses" (SPR)

## Install

You will need to have both `Go` (specifically [version 1.12](https://golang.org/dl/) or higher because we're using [Go modules](https://github.com/golang/go/wiki/Modules)) and the `make` programs installed on your computer. Assuming you do just type:

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Interface

_Please finish writing me..._

```
type StandardPlacesResult interface {
	Id() string
	ParentId() string
	Name() string
	Placetype() string
	Country() string
	Repo() string
	Path() string
	URI() string
	Latitude() float64
	Longitude() float64
	MinLatitude() float64
	MinLongitude() float64
	MaxLatitude() float64
	MaxLongitude() float64
	IsCurrent() flags.ExistentialFlag
	IsCeased() flags.ExistentialFlag
	IsDeprecated() flags.ExistentialFlag
	IsSuperseded() flags.ExistentialFlag
	IsSuperseding() flags.ExistentialFlag
	SupersededBy() []int64
	Supersedes() []int64
}
```

Flags are defined in the [go-whosonfirst-flags](https://github.com/whosonfirst/go-whosonfirst-flags) package.

## Background

_Please write me..._

* https://code.flickr.net/2008/08/19/standard-photos-response-apis-for-civilized-age/
* https://code.flickr.net/2008/08/25/api-responses-as-feeds/

## See also

* https://github.com/whosonfirst/go-whosonfirst-geojson-v2
* https://github.com/whosonfirst/go-whosonfirst-flags
