package uri

import (
	"fmt"
	_ "log"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// URI_REGEXP is the regular expression used to parse Who's On First URIs.
const URI_REGEXP string = `^(\d+)(?:\-alt(?:\-([a-zA-Z0-9_]+(?:\-[a-zA-Z0-9_]+(?:\-[a-zA-Z0-9_\-]+)?)?)))?(?:\.[^\.]+|\/)?$`

// re_uri is the internal *regexp.Regexp instance used to parse Who's On First URIs.
var re_uri *regexp.Regexp

func init() {
	re_uri = regexp.MustCompile(URI_REGEXP)
}

// IsAlternateGeometry returns a boolean value indicating whether 'path' is considered by an alternate geometry URI.
func IsAlternateGeometry(path string) (bool, error) {

	_, uri_args, err := ParseURI(path)

	if err != nil {
		return false, fmt.Errorf("Failed to parse '%s', %w", path, err)
	}

	return uri_args.IsAlternate, nil
}

// ParseURI will parse a Who's On First URI into its unique ID and any optional "alternate" geometry information.
func ParseURI(path string) (int64, *URIArgs, error) {

	abs_path, err := filepath.Abs(path)

	if err != nil {
		return -1, nil, fmt.Errorf("Failed to derive absolute path for %s, %w", path, err)
	}

	fname := filepath.Base(abs_path)

	match := re_uri.FindStringSubmatch(fname)

	// log.Println(fname, match)

	if len(match) == 0 {
		return -1, nil, fmt.Errorf("Unable to parse WOF ID for %s", path)
	}

	if len(match) < 2 {
		return -1, nil, fmt.Errorf("Unable to parse WOF ID for %s", path)
	}

	str_id := match[1]
	str_alt := match[2]

	wofid, err := strconv.ParseInt(str_id, 10, 64)

	if err != nil {
		return -1, nil, fmt.Errorf("Failed to parse %s, %w", str_id, err)
	}

	args := &URIArgs{
		IsAlternate: false,
	}

	if str_alt != "" {

		alt := strings.Split(str_alt, "-")

		alt_geom := &AltGeom{}

		switch len(alt) {
		case 1:
			alt_geom.Source = alt[0]
		case 2:
			alt_geom.Source = alt[0]
			alt_geom.Function = alt[1]
		default:
			alt_geom.Source = alt[0]
			alt_geom.Function = alt[1]
			alt_geom.Extras = alt[2:]
		}

		args.AltGeom = alt_geom
		args.IsAlternate = true
	}

	return wofid, args, nil
}

// ISWOFFile returns a boolean value indicating whether a path is a valid Who's On First URI.
func IsWOFFile(path string) (bool, error) {

	_, _, err := ParseURI(path)

	if err != nil {
		return false, nil
	}

	ext := filepath.Ext(path)

	if ext != ".geojson" {
		return false, nil
	}

	return true, nil
}

// ISAltFile returns a boolean value indicating whether a path is a valid Who's On First URI for an "alternate" geometry.
func IsAltFile(path string) (bool, error) {

	_, uri_args, err := ParseURI(path)

	if err != nil {
		return false, err
	}

	is_alt := uri_args.IsAlternate
	return is_alt, nil
}

// AltGeomFromPath parses a path and returns its *AltGeom instance if it is a valid Who's On First "alternate" geometry URI.
func AltGeomFromPath(path string) (*AltGeom, error) {

	_, uri_args, err := ParseURI(path)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	if !uri_args.IsAlternate {
		return nil, fmt.Errorf("%s is not an alternate geometry", path)
	}

	return uri_args.AltGeom, nil
}

// IdFromPath parses a path and return its unique Who's On First ID.
func IdFromPath(path string) (int64, error) {

	id, _, err := ParseURI(path)

	if err != nil {
		return 0, fmt.Errorf("Failed to parse URI, %w", err)
	}

	return id, nil
}

// RepoFromPath parses a path and if it is a valid whosonfirst-data Who's On First URI returns a GitHub repository name.
func WhosOnFirstDataRepoFromPath(path string) (string, error) {

	abs_path, err := filepath.Abs(path)

	if err != nil {
		return "", fmt.Errorf("Failed to derive absolute path for %s, %w", path, err)
	}

	wofid, err := IdFromPath(abs_path)

	if err != nil {
		return "", err
	}

	rel_path, err := Id2RelPath(wofid)

	if err != nil {
		return "", err
	}

	root_path := strings.Replace(abs_path, rel_path, "", 1)
	root_path = strings.TrimRight(root_path, "/")

	repo := ""

	for {

		base := filepath.Base(root_path)
		root_path = filepath.Dir(root_path)

		if strings.HasPrefix(base, "whosonfirst-data") {
			repo = base
			break
		}

		if root_path == "/" {
			break
		}

		if root_path == "" {
			break
		}
	}

	if repo == "" {
		return "", fmt.Errorf("Unable to determine repo from %s", path)
	}

	return repo, nil
}
