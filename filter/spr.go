package filter

import (
	"github.com/whosonfirst/go-whosonfirst-flags"
	"github.com/whosonfirst/go-whosonfirst-flags/existential"
	"github.com/whosonfirst/go-whosonfirst-flags/placetypes"
	_ "log"
	"strconv"
	"strings"
)

type SPRInputs struct {
	Placetypes    []string
	IsCurrent     []string
	IsCeased      []string
	IsDeprecated  []string
	IsSuperseded  []string
	IsSuperseding []string
}

type SPRFilter struct {
	Filter
	Placetypes  []flags.PlacetypeFlag
	Current     []flags.ExistentialFlag
	Deprecated  []flags.ExistentialFlag
	Ceased      []flags.ExistentialFlag
	Superseded  []flags.ExistentialFlag
	Superseding []flags.ExistentialFlag
}

func (f *SPRFilter) HasPlacetypes(fl flags.PlacetypeFlag) bool {

	for _, p := range f.Placetypes {

		if p.MatchesAny(fl) {
			return true
		}
	}

	return false
}

func (f *SPRFilter) IsCurrent(fl flags.ExistentialFlag) bool {

	for _, e := range f.Current {

		if e.MatchesAny(fl) {
			return true
		}
	}

	return false
}

func (f *SPRFilter) IsDeprecated(fl flags.ExistentialFlag) bool {

	for _, e := range f.Deprecated {

		if e.MatchesAny(fl) {
			return true
		}
	}

	return false
}

func (f *SPRFilter) IsCeased(fl flags.ExistentialFlag) bool {

	for _, e := range f.Ceased {

		if e.MatchesAny(fl) {
			return true
		}
	}

	return false
}

func (f *SPRFilter) IsSuperseded(fl flags.ExistentialFlag) bool {

	for _, e := range f.Superseded {

		if e.MatchesAny(fl) {
			return true
		}
	}

	return false
}

func (f *SPRFilter) IsSuperseding(fl flags.ExistentialFlag) bool {

	for _, e := range f.Superseding {

		if e.MatchesAny(fl) {
			return true
		}
	}

	return false
}

func NewSPRInputs() (*SPRInputs, error) {

	i := SPRInputs{
		Placetypes:    make([]string, 0),
		IsCurrent:     make([]string, 0),
		IsDeprecated:  make([]string, 0),
		IsCeased:      make([]string, 0),
		IsSuperseded:  make([]string, 0),
		IsSuperseding: make([]string, 0),
	}

	return &i, nil
}

func NewSPRFilter() (*SPRFilter, error) {

	null_pt, _ := placetypes.NewNullFlag()
	null_ex, _ := existential.NewNullFlag()

	col_pt := []flags.PlacetypeFlag{null_pt}
	col_ex := []flags.ExistentialFlag{null_ex}

	f := SPRFilter{
		Placetypes:  col_pt,
		Current:     col_ex,
		Deprecated:  col_ex,
		Ceased:      col_ex,
		Superseded:  col_ex,
		Superseding: col_ex,
	}

	return &f, nil
}

func NewSPRFilterFromInputs(inputs *SPRInputs) (Filter, error) {

	f, err := NewSPRFilter()

	if err != nil {
		return nil, err
	}

	if len(inputs.Placetypes) != 0 {

		possible, err := placetypeFlags(inputs.Placetypes)

		if err != nil {
			return nil, err
		}

		f.Placetypes = possible
	}

	if len(inputs.IsCurrent) != 0 {

		possible, err := existentialFlags(inputs.IsCurrent)

		if err != nil {
			return nil, err
		}

		f.Current = possible
	}

	if len(inputs.IsDeprecated) != 0 {

		possible, err := existentialFlags(inputs.IsDeprecated)

		if err != nil {
			return nil, err
		}

		f.Deprecated = possible
	}

	if len(inputs.IsCeased) != 0 {

		possible, err := existentialFlags(inputs.IsCeased)

		if err != nil {
			return nil, err
		}

		f.Ceased = possible
	}

	if len(inputs.IsSuperseded) != 0 {

		possible, err := existentialFlags(inputs.IsSuperseded)

		if err != nil {
			return nil, err
		}

		f.Superseded = possible
	}

	if len(inputs.IsSuperseding) != 0 {

		possible, err := existentialFlags(inputs.IsSuperseding)

		if err != nil {
			return nil, err
		}

		f.Superseding = possible
	}

	return f, nil
}

func placetypeFlags(inputs []string) ([]flags.PlacetypeFlag, error) {

	possible := make([]flags.PlacetypeFlag, 0)

	for _, test := range inputs {

		for _, pt := range strings.Split(test, ",") {

			pt = strings.Trim(pt, " ")

			fl, err := placetypes.NewPlacetypeFlag(pt)

			if err != nil {
				return nil, err
			}

			possible = append(possible, fl)
		}
	}

	return possible, nil
}

func existentialFlags(inputs []string) ([]flags.ExistentialFlag, error) {

	possible := make([]flags.ExistentialFlag, 0)

	for _, test := range inputs {

		for _, str_i := range strings.Split(test, ",") {

			str_i = strings.Trim(str_i, " ")

			i, err := strconv.ParseInt(str_i, 10, 64)

			if err != nil {
				return nil, err
			}

			fl, err := existential.NewKnownUnknownFlag(i)

			if err != nil {
				return nil, err
			}

			possible = append(possible, fl)
		}
	}

	return possible, nil
}
