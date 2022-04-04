package spr

import (
	"github.com/whosonfirst/go-whosonfirst-flags"
	"github.com/sfomuseum/go-edtf"	
)

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

type Pagination interface {
	Pages() int
	Page() int
	PerPage() int
	Total() int
	Cursor() string
	NextQuery() string
}

type StandardPlacesResults interface {
	Results() []StandardPlacesResult
}
