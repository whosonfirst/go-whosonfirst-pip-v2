package query

import (
	"errors"
	"regexp"
	"strings"
)

const SEP string = "="

type QueryFlags []*Query

func (m *QueryFlags) String() string {
	return ""
}

func (m *QueryFlags) Set(value string) error {

	parts := strings.Split(value, SEP)

	if len(parts) != 2 {
		return errors.New("Invalid query flag")
	}

	path := parts[0]
	str_match := parts[1]

	re, err := regexp.Compile(str_match)

	if err != nil {
		return err
	}

	q := &Query{
		Path:  path,
		Match: re,
	}

	*m = append(*m, q)
	return nil
}
