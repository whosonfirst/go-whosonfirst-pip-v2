package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	geojson_utils "github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	"github.com/whosonfirst/go-whosonfirst-pip/app"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	"github.com/whosonfirst/go-whosonfirst-pip/flags"
	"github.com/whosonfirst/go-whosonfirst-pip/utils"
	log "log"
	"os"
	"strconv"
	"strings"
)

func main() {

	fl, err := flags.CommonFlags()

	if err != nil {
		log.Fatal(err)
	}

	flags.Parse(fl)

	err = flags.ValidateCommonFlags(fl)

	if err != nil {
		log.Fatal(err)
	}

	pip, err := app.NewPIPApplication(fl)

	if err != nil {
		log.Fatal("Failed to create new PIP application, because", err)
	}

	pip_index, _ := flags.StringVar(fl, "index")
	pip_cache, _ := flags.StringVar(fl, "cache")
	mode, _ := flags.StringVar(fl, "mode")

	pip.Logger.Info("index is %s cache is %s mode is %s", pip_index, pip_cache, mode)

	err = pip.IndexPaths(fl.Args())

	if err != nil {
		pip.Logger.Fatal("Failed to index paths, because %s", err)
	}

	f, err := filter.NewSPRFilter()

	if err != nil {
		pip.Logger.Fatal("Failed to create SPR filter, because %s", err)
	}

	// ADD WAIT FOR INDEXER CODE HERE

	fmt.Println("ready to query")

	appindex := pip.Index

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {

		input := scanner.Text()
		pip.Logger.Status("# %s", input)

		parts := strings.Split(input, " ")

		if len(parts) == 0 {
			pip.Logger.Warning("Invalid input")
			continue
		}

		var command string

		switch parts[0] {

		case "candidates":
			command = parts[0]
		case "pip":
			command = parts[0]
		case "polyline":
			command = parts[0]
		default:
			pip.Logger.Warning("Invalid command")
			continue
		}

		var results interface{}

		if command == "pip" || command == "candidates" {

			str_lat := strings.Trim(parts[1], " ")
			str_lon := strings.Trim(parts[2], " ")

			lat, err := strconv.ParseFloat(str_lat, 64)

			if err != nil {
				pip.Logger.Warning("Invalid latitude, %s", err)
				continue
			}

			lon, err := strconv.ParseFloat(str_lon, 64)

			if err != nil {
				pip.Logger.Warning("Invalid longitude, %s", err)
				continue
			}

			c, err := geojson_utils.NewCoordinateFromLatLons(lat, lon)

			if err != nil {
				pip.Logger.Warning("Invalid latitude, longitude, %s", err)
				continue
			}

			if command == "pip" {

				intersects, err := appindex.GetIntersectsByCoord(c, f)

				if err != nil {
					pip.Logger.Warning("Unable to get intersects, because %s", err)
					continue
				}

				results = intersects

			} else {

				candidates, err := appindex.GetCandidatesByCoord(c)

				if err != nil {
					pip.Logger.Warning("Unable to get candidates, because %s", err)
					continue
				}

				results = candidates
			}

		} else if command == "polyline" {

			poly := parts[1]
			factor := 1.0e5

			if len(parts) > 2 {

				f, err := utils.StringPrecisionToFactor(parts[2])

				if err != nil {
					pip.Logger.Warning("Unable to parse precision because %s", err)
					continue
				}

				factor = f
			}

			path, err := utils.DecodePolyline(poly, factor)

			if err != nil {
				pip.Logger.Warning("Unable to decode polyline because %s", err)
				continue
			}

			intersects, err := appindex.GetIntersectsByPath(*path, f)

			if err != nil {
				pip.Logger.Warning("Unable to get candidates, because %s", err)
				continue
			}

			results = intersects

		} else {
			pip.Logger.Warning("Invalid command")
			continue
		}

		body, err := json.Marshal(results)

		if err != nil {
			pip.Logger.Warning("Failed to marshal results, because %s", err)
			continue
		}

		fmt.Println(string(body))
	}

	os.Exit(0)
}
