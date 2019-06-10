package tags

import (
	"errors"
	"fmt"
	"github.com/whosonfirst/go-rfc-5646"
	"github.com/whosonfirst/go-whosonfirst-names"
	"strings"
)

type LangTag struct {
	rfc5646.LanguageTag
	language   string
	extlang    string
	script     string
	region     string
	variant    string
	extension  string
	privateuse string
}

func (lt *LangTag) Language() string {
	return lt.language
}

func (lt *LangTag) ExtLang() string {
	return lt.extlang
}

func (lt *LangTag) Script() string {
	return lt.script
}

func (lt *LangTag) Region() string {
	return lt.region
}

func (lt *LangTag) Variant() string {
	return lt.variant
}

func (lt *LangTag) Extension() string {
	return lt.extension
}

func (lt *LangTag) PrivateUse() string {
	return lt.privateuse
}

func (lt *LangTag) String() string {

	possible := []string{
		lt.Language(),
		lt.ExtLang(),
		lt.Script(),
		lt.Region(),
		lt.Variant(),
		lt.Extension(),
		lt.PrivateUse(),
	}

	actual := make([]string, 0)

	for _, p := range possible {

		if p != "" {
			actual = append(actual, p)
		}
	}

	return strings.Join(actual, "_")
}

func NewLangTag(t string) (rfc5646.LanguageTag, error) {

	re := names.RE_LANGUAGETAG

	match := re.FindStringSubmatch(t)

	if len(match) == 0 {
		msg := fmt.Sprintf("Failed to parse language tag '%s'", t)
		return nil, errors.New(msg)
	}

	result := make(map[string]string)

	for i, name := range re.SubexpNames() {

		if i != 0 {
			result[name] = match[i]
		}
	}

	lt := LangTag{
		language:   result["language"],
		extlang:    result["extlang"],
		script:     result["script"],
		region:     result["region"],
		variant:    result["variant"],
		extension:  result["extension"],
		privateuse: result["privateuse"],
	}

	return &lt, nil
}
