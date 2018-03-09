package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-names/tags"
	"github.com/whosonfirst/go-whosonfirst-names/utils"	
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
		log.Println(utils.ToRFC5646(langtag.String()))
	}
}
