package feature

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	"io"
	"io/ioutil"
	"os"
)

// Feature

func LoadFeature(body []byte) (geojson.Feature, error) {

	is_wof := isWOF(body)
	is_alt := isAlt(body)

	if is_wof && is_alt {
		return NewWOFAltFeature(body)
	} else if is_wof {
		return NewWOFFeature(body)
	} else {
		return NewGeoJSONFeature(body)
	}
}

func LoadFeatureFromReader(fh io.Reader) (geojson.Feature, error) {

	body, err := UnmarshalFeatureFromReader(fh)

	if err != nil {
		return nil, err
	}

	return LoadFeature(body)
}

func LoadFeatureFromFile(path string) (geojson.Feature, error) {

	body, err := UnmarshalFeatureFromFile(path)

	if err != nil {
		return nil, err
	}

	return LoadFeature(body)
}

// WOF

func LoadWOFFeatureFromReader(fh io.Reader) (geojson.Feature, error) {

	body, err := UnmarshalFeatureFromReader(fh)

	if err != nil {
		return nil, err
	}

	return NewWOFFeature(body)
}

func LoadWOFFeatureFromFile(path string) (geojson.Feature, error) {

	body, err := UnmarshalFeatureFromFile(path)

	if err != nil {
		return nil, err
	}

	return NewWOFFeature(body)
}

func LoadWOFAltFeatureFromReader(fh io.Reader) (geojson.Feature, error) {

	body, err := UnmarshalFeatureFromReader(fh)

	if err != nil {
		return nil, err
	}

	return NewWOFAltFeature(body)
}

func LoadWOFAltFeatureFromFile(path string) (geojson.Feature, error) {

	body, err := UnmarshalFeatureFromFile(path)

	if err != nil {
		return nil, err
	}

	return NewWOFAltFeature(body)
}

// GeoJSON

func LoadGeoJSONFeatureFromReader(fh io.Reader) (geojson.Feature, error) {

	body, err := UnmarshalFeatureFromReader(fh)

	if err != nil {
		return nil, err
	}

	return NewGeoJSONFeature(body)
}

func LoadGeoJSONFeatureFromFile(path string) (geojson.Feature, error) {

	body, err := UnmarshalFeatureFromFile(path)

	if err != nil {
		return nil, err
	}

	return NewGeoJSONFeature(body)
}

func UnmarshalFeature(body []byte) ([]byte, error) {

	var stub interface{}
	err := json.Unmarshal(body, &stub)

	if err != nil {
		return nil, err
	}

	all := []string{
		"geometry",
		"geometry.type",
		"geometry.coordinates",
		"type",
	}

	err = utils.EnsureProperties(body, all)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func UnmarshalFeatureFromReader(fh io.Reader) ([]byte, error) {

	body, err := ioutil.ReadAll(fh)

	if err != nil {
		return nil, err
	}

	return UnmarshalFeature(body)
}

func UnmarshalFeatureFromFile(path string) ([]byte, error) {

	fh, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer fh.Close()

	return UnmarshalFeatureFromReader(fh)
}

func isWOF(body []byte) bool {
	wofid := gjson.GetBytes(body, "properties.wof:id")
	return wofid.Exists()
}

func isAlt(body []byte) bool {
	alt_label := gjson.GetBytes(body, "properties.src:alt_label")
	return alt_label.Exists()
}
