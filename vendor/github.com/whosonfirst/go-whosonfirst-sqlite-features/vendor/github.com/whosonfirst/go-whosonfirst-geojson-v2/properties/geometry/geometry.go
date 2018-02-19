package geometry

import (
	"errors"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
)

func ToString(f geojson.Feature) (string, error) {

	geom := gjson.GetBytes(f.Bytes(), "geometry")

	if !geom.Exists() {
		return "", errors.New("Missing geometry property")
	}

	return geom.Raw, nil
}

func Type(f geojson.Feature) string {

	possible := []string{
		"geometry.type",
	}

	return utils.StringProperty(f.Bytes(), possible, "unknown")
}
