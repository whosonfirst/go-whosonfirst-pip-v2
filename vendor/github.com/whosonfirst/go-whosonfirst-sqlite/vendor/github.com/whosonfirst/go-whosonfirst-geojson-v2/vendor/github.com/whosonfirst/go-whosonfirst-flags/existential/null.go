package existential

import (
	"github.com/whosonfirst/go-whosonfirst-flags"
	"strconv"
)

type NullFlag struct {
	flags.ExistentialFlag
}

func NewNullFlag() (flags.ExistentialFlag, error) {

	n := NullFlag{}
	return &n, nil
}

func (f *NullFlag) StringFlag() string {
	return strconv.FormatInt(f.Flag(), 10)
}

func (f *NullFlag) Flag() int64 {
	return -1
}

func (f *NullFlag) IsTrue() bool {
	return false
}

func (f *NullFlag) IsFalse() bool {
	return false
}

func (f *NullFlag) IsKnown() bool {
	return false
}

func (f *NullFlag) MatchesAny(others ...flags.ExistentialFlag) bool {
	return true
}

func (f *NullFlag) MatchesAll(others ...flags.ExistentialFlag) bool {
	return true
}

func (f *NullFlag) String() string {
	return "NULL"
}
