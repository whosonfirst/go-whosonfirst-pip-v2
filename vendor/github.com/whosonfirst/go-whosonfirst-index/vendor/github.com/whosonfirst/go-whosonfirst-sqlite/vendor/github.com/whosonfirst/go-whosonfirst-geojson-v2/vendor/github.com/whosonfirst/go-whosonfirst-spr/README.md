# go-whosonfirst-spr

Go tools for working Who's On First "standard places responses" (SPR)

## Install

You will need to have both `Go` (specifically a version of Go more recent than 1.6 so let's just assume you need [Go 1.8](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make bin
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Important

Too soon. Way way too soon. Move along. Nothing should be considered "stable" yet. If you want to follow along, please consult:

https://github.com/whosonfirst/go-whosonfirst-spr/issues/1

## Interface

_Please finish writing me..._

```
type StandardPlacesResult interface {
	Id() int64
	ParentId() int64
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

## Background

_Please write me..._

* https://code.flickr.net/2008/08/19/standard-photos-response-apis-for-civilized-age/
* https://code.flickr.net/2008/08/25/api-responses-as-feeds/

## See also

* https://github.com/whosonfirst/go-whosonfirst-geojson-v2
