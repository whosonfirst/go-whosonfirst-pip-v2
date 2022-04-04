package existential

import (
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-flags"
	"strconv"
)

type KnownUnknownFlag struct {
	flags.ExistentialFlag
	flag       int64 // https://github.com/whosonfirst/go-whosonfirst-flags/issues/2
	status     bool
	confidence bool
}

func NewKnownUnknownFlagsArray(values ...int64) ([]flags.ExistentialFlag, error) {

	existential_flags := make([]flags.ExistentialFlag, 0)

	for _, v := range values {

		fl, err := NewKnownUnknownFlag(v)

		if err != nil {
			return nil, err
		}

		existential_flags = append(existential_flags, fl)
	}

	return existential_flags, nil
}

func NewKnownUnknownFlag(i int64) (flags.ExistentialFlag, error) {

	var status bool
	var confidence bool

	switch i {
	case 0:
		status = false
		confidence = true
	case 1:
		status = true
		confidence = true
	default:
		i = -1 // just in case someone passes us garbage
		status = false
		confidence = false
	}

	f := KnownUnknownFlag{
		flag:       i,
		status:     status,
		confidence: confidence,
	}

	return &f, nil
}

func (f *KnownUnknownFlag) StringFlag() string {
	return strconv.FormatInt(f.Flag(), 10)
}

func (f *KnownUnknownFlag) Flag() int64 {
	return f.flag
}

func (f *KnownUnknownFlag) IsTrue() bool {
	return f.status == true
}

func (f *KnownUnknownFlag) IsFalse() bool {
	return f.status == false
}

func (f *KnownUnknownFlag) IsKnown() bool {
	return f.confidence
}

func (f *KnownUnknownFlag) MatchesAny(others ...flags.ExistentialFlag) bool {

	for _, o := range others {
		if f.Flag() == o.Flag() {
			return true
		}
	}

	return false
}

func (f *KnownUnknownFlag) MatchesAll(others ...flags.ExistentialFlag) bool {

	matches := 0

	for _, o := range others {
		if f.Flag() == o.Flag() {
			matches += 1
		}
	}

	if matches == len(others) {
		return true
	}

	return false
}

func (f *KnownUnknownFlag) String() string {
	return fmt.Sprintf("FLAG %d IS TRUE %t IS FALSE %t IS  KNOWN %t", f.flag, f.IsTrue(), f.IsFalse(), f.IsKnown())
}
