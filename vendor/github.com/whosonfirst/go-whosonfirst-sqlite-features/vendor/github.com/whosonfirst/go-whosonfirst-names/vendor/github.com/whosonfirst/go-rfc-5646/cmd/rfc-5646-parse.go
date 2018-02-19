package main

import (
	"flag"
	"github.com/whosonfirst/go-rfc-5646/tags"
	"log"
)

func main() {

	flag.Parse()

	for _, raw := range flag.Args() {

		log.Println(raw)
		langtag, err := tags.NewLangTag(raw)

		if err != nil {
			log.Fatal(err)
		}

		log.Println(langtag.String())
	}
}
