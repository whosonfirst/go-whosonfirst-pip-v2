package main

import (
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/geometry"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	"log"
)

func main() {

	show_geom := flag.Bool("geom", false, "...")

	flag.Parse()
	args := flag.Args()

	for _, path := range args {

		f, err := feature.LoadFeatureFromFile(path)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("# %s\n", path)

		fmt.Printf("ID is %s\n", f.Id())
		fmt.Printf("WOF ID is %d\n", whosonfirst.Id(f))
		fmt.Printf("Name is %s\n", f.Name())
		fmt.Printf("Placetype is %s\n", f.Placetype())

		coord, _ := utils.NewCoordinateFromLatLons(0.0, 0.0)
		contains, _ := f.ContainsCoord(coord)

		fmt.Printf("Contains %v %t\n", coord, contains)

		bboxes, _ := f.BoundingBoxes()

		fmt.Printf("Count boxes %d\n", len(bboxes.Bounds()))
		fmt.Printf("MBR %s\n", bboxes.MBR())

		wof, err := feature.LoadWOFFeatureFromFile(path)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("WOF repo is %s\n", whosonfirst.Repo(wof))

		str_geom, err := geometry.ToString(wof)

		if err != nil {
			log.Fatal(err)
		}

		if *show_geom {
			fmt.Println(str_geom)
		}
	}

}
