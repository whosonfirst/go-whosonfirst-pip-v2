package placetypes

import (
	"encoding/json"
	"errors"
	"github.com/whosonfirst/go-whosonfirst-placetypes/placetypes"
	"log"
	"strconv"
)

type WOFPlacetypeName struct {
	Lang string `json:"language"`
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type WOFPlacetypeAltNames map[string][]string

type WOFPlacetype struct {
	Id     int64   `json:"id"`
	Name   string  `json:"name"`
	Role   string  `json:"role"`
	Parent []int64 `json:"parent"`
	// AltNames []WOFPlacetypeAltNames		`json:"names"`
}

type WOFPlacetypeSpecification map[string]WOFPlacetype

var specification *WOFPlacetypeSpecification

func init() {

	var err error

	specification, err = Spec()

	if err != nil {
		log.Fatal("Failed to parse specification", err)
	}
}

func Spec() (*WOFPlacetypeSpecification, error) {

	var spec WOFPlacetypeSpecification
	err := json.Unmarshal([]byte(placetypes.Specification), &spec)

	if err != nil {
		return nil, err
	}

	return &spec, nil
}

func IsValidPlacetype(name string) bool {

	for _, pt := range *specification {

		if pt.Name == name {
			return true
		}
	}

	return false
}

func IsValidPlacetypeId(id int64) bool {

	for str_id, _ := range *specification {

		pt_id, err := strconv.Atoi(str_id)

		if err != nil {
			continue
		}

		pt_id64 := int64(pt_id)

		if pt_id64 == id {
			return true
		}
	}

	return false
}

func GetPlacetypeByName(name string) (*WOFPlacetype, error) {

	for str_id, pt := range *specification {

		if pt.Name == name {

			pt_id, err := strconv.Atoi(str_id)

			if err != nil {
				continue
			}

			pt_id64 := int64(pt_id)

			pt.Id = pt_id64
			return &pt, nil
		}
	}

	return nil, errors.New("Invalid placetype")
}

func GetPlacetypeById(id int64) (*WOFPlacetype, error) {

	for str_id, pt := range *specification {

		pt_id, err := strconv.Atoi(str_id)

		if err != nil {
			continue
		}

		pt_id64 := int64(pt_id)

		if pt_id64 == id {
			pt.Id = pt_id64
			return &pt, nil
		}
	}

	return nil, errors.New("Invalid placetype")
}
