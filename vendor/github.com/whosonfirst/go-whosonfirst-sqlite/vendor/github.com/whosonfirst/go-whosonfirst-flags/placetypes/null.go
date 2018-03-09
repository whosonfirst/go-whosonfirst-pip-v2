package placetypes

import (
	"github.com/whosonfirst/go-whosonfirst-flags"
)

type NullFlag struct {
	flags.PlacetypeFlag
}

func NewNullFlag() (*NullFlag, error) {

	f := NullFlag{}
	return &f, nil
}

func (f *NullFlag) MatchesAny(others ...flags.PlacetypeFlag) bool {
	return true
}

func (f *NullFlag) MatchesAll(others ...flags.PlacetypeFlag) bool {
	return true
}

func (f *NullFlag) Placetype() string {
	return ""
}

func (f *NullFlag) String() string {
	return "NULL"
}
