package sources

import (
	"encoding/json"
	"errors"
	"github.com/whosonfirst/go-whosonfirst-sources/sources"
	"log"
)

type WOFSource struct {
	Id          int    `json:"id"`
	Fullname    string `json:"fullname"`
	Name        string `json:"name"`
	Prefix      string `json:"prefix"`
	Key         string `json:"key"`
	URL         string `json:"url"`
	License     string `json:"license"`
	Description string `json:"description"`
}

type WOFSourceSpecification map[string]WOFSource

var specification *WOFSourceSpecification

func init() {

	var err error

	specification, err = Spec()

	if err != nil {
		log.Fatal("Failed to parse specification", err)
	}
}

func Spec() (*WOFSourceSpecification, error) {

	var spec WOFSourceSpecification
	err := json.Unmarshal([]byte(sources.Specification), &spec)

	if err != nil {
		return nil, err
	}

	return &spec, nil
}

func IsValidSource(source string) bool {

	for _, details := range *specification {

		if details.Name == source {
			return true
		}
	}

	return false
}

func IsValidSourceId(source_id int) bool {

	for _, details := range *specification {

		if details.Id == source_id {
			return true
		}
	}

	return false
}

func GetSourceByName(source string) (*WOFSource, error) {

	for _, details := range *specification {

		if details.Name == source {
			return &details, nil
		}
	}

	return nil, errors.New("Invalid source")
}

func GetSourceById(source_id int) (*WOFSource, error) {

	for _, details := range *specification {

		if details.Id == source_id {
			return &details, nil
		}
	}

	return nil, errors.New("Invalid source")
}
