package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
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

		is_current, _ := whosonfirst.IsCurrent(f)
		is_current_raw := utils.StringProperty(f.Bytes(), []string{"properties.mz:is_current"}, "")

		log.Printf("is current: %s raw:%s\n", is_current, is_current_raw)

		is_deprecated, _ := whosonfirst.IsDeprecated(f)
		is_deprecated_raw := utils.StringProperty(f.Bytes(), []string{"properties.edtf:deprecated"}, "")

		log.Printf("is deprecated:%s raw:%s\n", is_deprecated, is_deprecated_raw)

		is_ceased, _ := whosonfirst.IsCeased(f)
		is_ceased_raw := utils.StringProperty(f.Bytes(), []string{"properties.edtf:cessation"}, "")

		log.Printf("is ceased:%s raw:%s\n", is_ceased, is_ceased_raw)

		is_superseded, _ := whosonfirst.IsSuperseded(f)
		is_superseded_raw := utils.StringProperty(f.Bytes(), []string{"properties.wof:superseded_by"}, "")

		log.Printf("is superseded:%s raw:%s\n", is_superseded, is_superseded_raw)

		is_superseding, _ := whosonfirst.IsSuperseding(f)
		is_superseding_raw := utils.StringProperty(f.Bytes(), []string{"properties.wof:supersedes"}, "")

		log.Printf("is superseding:%s raw:%s\n", is_superseding, is_superseding_raw)

	}

}
