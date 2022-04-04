package feature

import (
	"fmt"
	"github.com/sfomuseum/go-edtf"
)

var deprecated map[string]string

func init() {

	deprecated = map[string]string{
		edtf.OPEN_2012:        edtf.OPEN,
		edtf.UNSPECIFIED_2012: edtf.UNSPECIFIED,
	}

}

func isDeprecatedEDTF(edtf_str string) bool {

	for test, _ := range deprecated {

		if edtf_str == test {
			return true
		}
	}

	return false
}

func replaceDeprecatedEDTF(old string) (string, error) {

	new, ok := deprecated[old]

	if !ok {
		err := fmt.Errorf("Unknown or unsupported EDTF string '%s' : %v", old, deprecated)
		return "", err
	}

	return new, nil
}
