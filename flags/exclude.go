package flags

// this is invoked/used in flags/flags.go and app/indexer.go but for the life
// of me I can't figure out how to make the code below implement the
// correct inferface wah wah so that flag.Lookup("exclude").Value returns
// something we can loop over... so instead we just strings.Split() on
// flag.Lookup("exclude").String() which is dumb but works...
// (20180301/thisisaaronland)

import (
	"strings"
)

type Exclude []string

func (e *Exclude) String() string {
	return strings.Join(*e, " ")
}

func (e *Exclude) Set(value string) error {
	*e = append(*e, value)
	return nil
}
