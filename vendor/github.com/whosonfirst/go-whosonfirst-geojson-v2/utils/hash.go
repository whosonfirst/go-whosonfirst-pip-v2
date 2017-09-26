package utils

import (
	"errors"
	"github.com/mmcloughlin/geohash"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-hash"
)

func GeohashFeature(f geojson.Feature) (string, error) {

	bboxes, err := f.BoundingBoxes()

	if err != nil {
		return "", err
	}

	mbr := bboxes.MBR()
	center := mbr.Center()

	lat := center.Y
	lon := center.X

	gh := geohash.Encode(lat, lon)
	return gh, nil
}

func HashFeature(f geojson.Feature) (string, error) {

	return "", errors.New("This is not ready to use yet")

	// what we want is for the output of (b) to be the same as (a)
	// (20170801/thisisaaronland)

	// hashing the file (a)

	/*
		h, err := hash.NewWOFHash()

		if err != nil {
			log.Fatal(err)
		}

		file_hash, err := h.HashFile(path)
	*/

	// hashing the feature (b)
	// github.com/whosonfirst/go-whosonfirst-export

	/*

		        e, err := export.ExportFeature(f.Bytes())

			if err != nil {
				return "", err
			}

			h, err := hash.NewWOFHash()

			if err != nil {
				return "", err
			}

			return h.HashFromJSON(e)

	*/
}

// this causes an import loop so we're just going to leave it
// here as a reference for now... (20170801/thisisaaronland)
// HashGeometryForFeature(f geojson.Feature) (string, error)
// geom, err := geometry.ToString(f)
// return HashGeometry([]byte(geom))

func HashGeometry(geom []byte) (string, error) {

	h, err := hash.NewWOFHash()

	if err != nil {
		return "", err
	}

	return h.HashFromJSON(geom)
}
