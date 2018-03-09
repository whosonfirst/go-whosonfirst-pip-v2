package main

import (
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-sources"
)

func main() {

	fmt.Println(sources.IsValidSource("sfac"))

	src, _ := sources.GetSourceByName("mapzen")
	fmt.Println(src.License)
}
