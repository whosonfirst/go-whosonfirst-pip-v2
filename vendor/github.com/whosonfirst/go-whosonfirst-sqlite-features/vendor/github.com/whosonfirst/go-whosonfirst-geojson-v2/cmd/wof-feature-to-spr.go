package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"log"
)

func main() {

	flag.Parse()

	for _, path := range flag.Args() {

		f, err := feature.LoadFeatureFromFile(path)

		if err != nil {
			log.Fatal(err)
		}

		s, err := f.SPR()

		if err != nil {
			log.Fatal(err)
		}

		body, err := json.Marshal(s)
		fmt.Println(string(body))
	}
}
