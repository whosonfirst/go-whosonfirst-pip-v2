# go-whosonfirst-spr

Go package for the Who's On First "standard places responses" (SPR) interface.

## Description

The `StandardPlacesResult` (SPR) interface defines the _minimum_ set of methods that a system working with a collection of Who's On First (WOF) must implement for any given record. Not all records are the same so the SPR interface is meant to serve as a baseline for common data that describes every record.

The `StandardPlacesResults` takes the Flickr [standard photo response](https://code.flickr.net/2008/08/19/standard-photos-response-apis-for-civilized-age) as its inspiration which was designed to be the minimum amount of information about a Flickr photo necessary to display that photo with proper attribution and a link back to the photo page itself. The `StandardPlacesResults` aims to achieve the same thing for WOF records.

Being a [Go language interface type](https://www.alexedwards.net/blog/interfaces-explained) the SPR is _not_ designed as a data exchange method. Any given implementation of the SPR _may_ allow its internal data to be exported or serialized (for example, as JSON) but this is not a requirement.

For a concrete example of a package that implements the `SPR` have a look at the [go-whosonfirst-sqlite-spr](https://github.com/whosonfirst/go-whosonfirst-sqlite-spr) package.

## Usage

```
import "github.com/whosonfirst/go-whosonfirst-spr/v2"
```

## Interface

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
	Inception() *edtf.EDTFDate
	Cessation() *edtf.EDTFDate	
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
	BelongsTo() []int64
	LastModified() int64
}
```

### Notes

* The `Id()` and `ParentId()` methods return `string` (rather than `int64`) values to account for non-WOF GeoJSON documents that are consumed by the `whosonfirst/go-whosonfirst-geojson-v2` package.

* Flags are defined in the [go-whosonfirst-flags](https://github.com/whosonfirst/go-whosonfirst-flags) package.

* The `Path()` method is expected to return a relative URI. The `URI` method is expected to return a fully-qualified URI. These two methods are confusing and that confusion should be addressed.

## See also

* https://github.com/whosonfirst/go-whosonfirst-geojson-v2
* https://github.com/whosonfirst/go-whosonfirst-flags
* https://github.com/whosonfirst/go-whosonfirst-sqlite-spr
* https://github.com/sfomuseum/go-edtf

### Related

* https://code.flickr.net/2008/08/19/standard-photos-response-apis-for-civilized-age/
* https://code.flickr.net/2008/08/25/api-responses-as-feeds/
