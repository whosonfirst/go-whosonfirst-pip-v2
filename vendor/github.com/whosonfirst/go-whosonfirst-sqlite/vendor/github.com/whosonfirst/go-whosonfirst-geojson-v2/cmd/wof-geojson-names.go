package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
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

		names_map := whosonfirst.Names(f)

		for tag, names := range names_map {

			for _, n := range names {

				log.Printf("%s %s %s\n", f.Id(), tag, n)
			}
		}
	}
}
