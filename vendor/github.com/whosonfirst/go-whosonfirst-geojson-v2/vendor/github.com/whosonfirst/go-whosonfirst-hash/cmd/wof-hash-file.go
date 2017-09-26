package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-hash"
	"log"
)

func main() {

	flag.Parse()

	algo := "md5"
	h, err := hash.NewHash(algo)

	if err != nil {
		log.Fatal(err)
	}

	for _, path := range flag.Args() {

		hashed, err := h.HashFile(path)

		if err != nil {
			log.Fatal(err)
		}

		log.Println(path, hashed)
	}
}
