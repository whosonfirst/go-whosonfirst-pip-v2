package main

import (
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/geometry"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	"github.com/whosonfirst/go-whosonfirst-hash"
	"log"
)

func main() {

	flag.Parse()
	args := flag.Args()

	for _, path := range args {

		f, err := feature.LoadWOFFeatureFromFile(path)

		if err != nil {
			log.Fatal(err)
		}

		h, err := hash.NewWOFHash()

		if err != nil {
			log.Fatal(err)
		}

		file_hash, err := h.HashFile(path)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("file hash is %s\n", file_hash)

		feature_hash, err := utils.HashFeature(f)

		if err != nil {
			fmt.Printf("failed to generate feature hash because %s\n", err)
		} else {
			fmt.Printf("feature hash is %s\n", feature_hash)
		}

		str_geom, err := geometry.ToString(f)

		if err != nil {
			log.Fatal(err)
		}

		geom_hash, err := utils.HashGeometry([]byte(str_geom))

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("geometry hash is %s\n", geom_hash)
	}

}
