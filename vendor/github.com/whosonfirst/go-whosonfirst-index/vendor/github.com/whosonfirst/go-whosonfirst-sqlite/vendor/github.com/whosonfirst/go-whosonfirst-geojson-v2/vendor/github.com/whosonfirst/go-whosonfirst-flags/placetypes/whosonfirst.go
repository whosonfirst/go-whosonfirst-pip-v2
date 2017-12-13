package placetypes

import (
	"github.com/whosonfirst/go-whosonfirst-flags"
	wof "github.com/whosonfirst/go-whosonfirst-placetypes"
)

type PlacetypeFlag struct {
	flags.PlacetypeFlag
	pt *wof.WOFPlacetype
}

func NewPlacetypeFlag(name string) (flags.PlacetypeFlag, error) {

	pt, err := wof.GetPlacetypeByName(name)

	if err != nil {
		return nil, err
	}

	f := PlacetypeFlag{
		pt: pt,
	}

	return &f, nil
}

func (f *PlacetypeFlag) MatchesAny(others ...flags.PlacetypeFlag) bool {

	for _, o := range others {

		if f.Placetype() == o.Placetype() {
			return true
		}

	}

	return false
}

func (f *PlacetypeFlag) MatchesAll(others ...flags.PlacetypeFlag) bool {

	matches := 0

	for _, o := range others {

		if f.Placetype() == o.Placetype() {
			matches += 1
		}

	}

	if matches == len(others) {
		return true
	}

	return false
}

func (f *PlacetypeFlag) Placetype() string {
	return f.pt.Name
}

func (f *PlacetypeFlag) String() string {
	return f.Placetype()
}
