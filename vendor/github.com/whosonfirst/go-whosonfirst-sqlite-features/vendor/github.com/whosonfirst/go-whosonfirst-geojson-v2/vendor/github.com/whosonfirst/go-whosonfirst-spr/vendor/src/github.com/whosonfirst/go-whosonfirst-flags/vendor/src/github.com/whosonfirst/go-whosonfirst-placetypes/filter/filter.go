package filter

import (
	"errors"
	"github.com/whosonfirst/go-whosonfirst-placetypes"
)

type PlacetypesFilter struct {
	required  map[string]*placetypes.WOFPlacetype
	forbidden map[string]*placetypes.WOFPlacetype
}

func NewPlacetypesFilter(include []string, include_roles []string, exclude []string) (*PlacetypesFilter, error) {

	required := make(map[string]*placetypes.WOFPlacetype)
	forbidden := make(map[string]*placetypes.WOFPlacetype)

	for _, p := range include {

		_, ok := required[p]

		if ok {
			continue
		}

		pt, err := placetypes.GetPlacetypeByName(p)

		if err != nil {
			return nil, err
		}

		required[p] = pt
	}

	if len(include_roles) > 0 {
		return nil, errors.New("included roles are not supported yet")
	}

	/*
		for _, p := range include_roles {

		}
	*/

	for _, p := range exclude {

		_, ok := forbidden[p]

		if ok {
			continue
		}

		pt, err := placetypes.GetPlacetypeByName(p)

		if err != nil {
			return nil, err
		}

		forbidden[p] = pt
	}

	f := PlacetypesFilter{
		required:  required,
		forbidden: forbidden,
	}

	return &f, nil
}

func (f *PlacetypesFilter) AllowFromString(pt_str string) (bool, error) {

	pt, err := placetypes.GetPlacetypeByName(pt_str)

	if err != nil {
		return false, err
	}

	return f.Allow(pt)
}

func (f *PlacetypesFilter) Allow(pt *placetypes.WOFPlacetype) (bool, error) {

	if len(f.forbidden) > 0 {

		_, ok := f.forbidden[pt.Name]

		if ok {
			return false, nil
		}

	}

	if len(f.required) > 0 {

		_, ok := f.required[pt.Name]

		if !ok {
			return false, nil
		}
	}

	return true, nil
}
