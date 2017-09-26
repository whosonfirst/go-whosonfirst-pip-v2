package main

/*

./bin/wof-geojson-pip -latitude 45.593352 -longitude -73.513992 /usr/local/data/whosonfirst-data/data/856/330/41/85633041.geojson
2017/08/21 22:16:52 /usr/local/data/whosonfirst-data/data/856/330/41/85633041.geojson true

./bin/wof-geojson-pip -latitude 45.593352 -longitude -73.513992 -verbose /usr/local/data/whosonfirst-data/data/856/330/41/85633041.geojson
2017/08/21 22:17:22 /usr/local/data/whosonfirst-data/data/856/330/41/85633041.geojson true
2017/08/21 22:17:22 /usr/local/data/whosonfirst-data/data/856/330/41/85633041.geojson 0 false
2017/08/21 22:17:22 /usr/local/data/whosonfirst-data/data/856/330/41/85633041.geojson 1 false
2017/08/21 22:17:22 /usr/local/data/whosonfirst-data/data/856/330/41/85633041.geojson 2 false
2017/08/21 22:17:22 /usr/local/data/whosonfirst-data/data/856/330/41/85633041.geojson 3 false
2017/08/21 22:17:22 /usr/local/data/whosonfirst-data/data/856/330/41/85633041.geojson 4 false
2017/08/21 22:17:22 /usr/local/data/whosonfirst-data/data/856/330/41/85633041.geojson 5 false
2017/08/21 22:17:22 /usr/local/data/whosonfirst-data/data/856/330/41/85633041.geojson 6 false
2017/08/21 22:17:22 /usr/local/data/whosonfirst-data/data/856/330/41/85633041.geojson 7 false
2017/08/21 22:17:22 /usr/local/data/whosonfirst-data/data/856/330/41/85633041.geojson 8 false
2017/08/21 22:17:22 /usr/local/data/whosonfirst-data/data/856/330/41/85633041.geojson 9 true
2017/08/21 22:17:22 /usr/local/data/whosonfirst-data/data/856/330/41/85633041.geojson 10 false
... and so on

./bin/wof-geojson-intersects -point 37.821,-122.2259 /usr/local/data/whosonfirst-data/data/859/218/81/85921881.geojson /usr/local/data/whosonfirst-data/data/859/218/77/85921877.geojson
2017/08/22 16:03:59 /usr/local/data/whosonfirst-data/data/859/218/81/85921881.geojson false	# oakland (CA)
2017/08/22 16:03:59 /usr/local/data/whosonfirst-data/data/859/218/77/85921877.geojson true	# piedmont (CA)

*/

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	"log"
	"strconv"
	"strings"
)

func main() {

	var lat = flag.Float64("latitude", 0.0, "...")
	var lon = flag.Float64("longitude", 0.0, "...")
	var point = flag.String("point", "", "")
	verbose := flag.Bool("verbose", false, "...")

	flag.Parse()

	if *point != "" {

		parts := strings.Split(*point, ",")

		if len(parts) != 2 {
			log.Fatal("Can not parse point")
		}

		str_lat := strings.Trim(parts[0], " ")
		str_lon := strings.Trim(parts[1], " ")

		fl_lat, err := strconv.ParseFloat(str_lat, 64)

		if err != nil {
			log.Fatal(err)
		}

		fl_lon, err := strconv.ParseFloat(str_lon, 64)

		if err != nil {
			log.Fatal(err)
		}

		*lat = fl_lat
		*lon = fl_lon
	}

	for _, path := range flag.Args() {

		f, err := feature.LoadWOFFeatureFromFile(path)

		if err != nil {
			log.Fatal(err)
		}

		coord, err := utils.NewCoordinateFromLatLons(*lat, *lon)

		if err != nil {
			log.Fatal(err)
		}

		contained, err := f.ContainsCoord(coord)

		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%s %t\n", path, contained)

		if !*verbose {
			continue
		}

		polys, err := f.Polygons()

		if err != nil {
			log.Fatal(err)
		}

		for i, p := range polys {

			poly_contained := p.ContainsCoord(coord)

			if err != nil {
				log.Fatal(err)
			}

			log.Printf("%s %d %t\n", path, i, poly_contained)
		}
	}

}
