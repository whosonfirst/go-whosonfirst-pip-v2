package filter

import (
	"errors"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-flags"
	"github.com/whosonfirst/go-whosonfirst-flags/placetypes"
	"github.com/whosonfirst/go-whosonfirst-spr"
	"log"
)

type Filter interface {
	HasPlacetypes(flags.PlacetypeFlag) bool
	IsCurrent(flags.ExistentialFlag) bool
	IsDeprecated(flags.ExistentialFlag) bool
	IsCeased(flags.ExistentialFlag) bool
	IsSuperseded(flags.ExistentialFlag) bool
	IsSuperseding(flags.ExistentialFlag) bool
}

func FilterSPR(filters Filter, s spr.StandardPlacesResult) error {

	var ok bool

	pf, err := placetypes.NewPlacetypeFlag(s.Placetype())

	if err != nil {
		msg := fmt.Sprintf("Unable to parse placetype (%s) for ID %s, because '%s' - skipping placetype filters", s.Placetype(), s.Id(), err)
		log.Println(msg)
	} else {

		ok = filters.HasPlacetypes(pf)

		if !ok {
			return errors.New("Failed 'placetype' test")
		}
	}

	ok = filters.IsCurrent(s.IsCurrent())

	if !ok {
		return errors.New("Failed 'is current' test")
	}

	ok = filters.IsDeprecated(s.IsDeprecated())

	if !ok {
		return errors.New("Failed 'is deprecated' test")
	}

	ok = filters.IsCeased(s.IsCeased())

	if !ok {
		return errors.New("Failed 'is ceased' test")
	}

	ok = filters.IsSuperseded(s.IsSuperseded())

	if !ok {
		return errors.New("Failed 'is superseded' test")
	}

	ok = filters.IsSuperseding(s.IsSuperseding())

	if !ok {
		return errors.New("Failed 'is superseding' test")
	}

	return nil
}
