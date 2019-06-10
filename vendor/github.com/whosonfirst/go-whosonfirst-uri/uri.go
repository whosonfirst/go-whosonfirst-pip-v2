package uri

import (
	"errors"
	"github.com/whosonfirst/go-whosonfirst-sources"
	_ "log"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type URIArgs struct {
	// PLEASE UPDATE THIS TO USE/EXPECT AN *AltGeom KTHXBYE (20190501/thisisaaronland)
	Alternate bool
	Source    string
	Function  string
	Extras    []string
	Strict    bool
}

type AltGeom struct {
	Source   string
	Function string
	Extras   []string
}

func (a *AltGeom) String() string {

	parts := []string{
		a.Source,
	}

	if a.Function != "" {
		parts = append(parts, a.Function)
	}

	for _, ex := range a.Extras {
		parts = append(parts, ex)
	}

	return strings.Join(parts, "-")
}

func NewDefaultURIArgs() *URIArgs {

	u := URIArgs{
		Alternate: false,
		Source:    "",
		Function:  "",
		Extras:    make([]string, 0),
		Strict:    false,
	}

	return &u
}

func NewAlternateURIArgs(source string, function string, extras ...string) *URIArgs {

	u := URIArgs{
		Alternate: true,
		Source:    source,
		Function:  function,
		Extras:    extras,
		Strict:    false,
	}

	return &u
}

// See also: https://github.com/whosonfirst/whosonfirst-cookbook/blob/master/how_to/creating_alt_geometries.md

func Id2Fname(id int64, args ...*URIArgs) (string, error) {

	str_id := strconv.FormatInt(id, 10)
	parts := []string{str_id}

	if len(args) == 1 {

		uri_args := args[0]

		if uri_args.Alternate {

			if uri_args.Source == "" && uri_args.Strict {
				return "", errors.New("Missing source argument for alternate geometry")
			}

			if uri_args.Source == "" {
				uri_args.Source = "unknown"

			}

			if uri_args.Strict && !sources.IsValidSource(uri_args.Source) {
				return "", errors.New("Invalid or unknown source argument for alternate geometry")
			}

			parts = append(parts, "alt")
			parts = append(parts, uri_args.Source)

			if uri_args.Function != "" {
				parts = append(parts, uri_args.Function)
			}

			for _, e := range uri_args.Extras {
				parts = append(parts, e)
			}
		}

	}

	str_parts := strings.Join(parts, "-")

	fname := str_parts + ".geojson"
	return fname, nil
}

func Id2Path(id int64) (string, error) {

	parts := []string{}
	input := strconv.FormatInt(id, 10)

	for len(input) > 3 {

		chunk := input[0:3]
		input = input[3:]
		parts = append(parts, chunk)
	}

	if len(input) > 0 {
		parts = append(parts, input)
	}

	path := filepath.Join(parts...)
	return path, nil
}

func Id2RelPath(id int64, args ...*URIArgs) (string, error) {

	fname, err := Id2Fname(id, args...)

	if err != nil {
		return "", err
	}

	root, err := Id2Path(id)

	if err != nil {
		return "", err
	}

	rel_path := filepath.Join(root, fname)
	return rel_path, nil
}

func Id2AbsPath(root string, id int64, args ...*URIArgs) (string, error) {

	rel, err := Id2RelPath(id, args...)

	if err != nil {
		return "", err
	}

	var abs_path string

	// because filepath.Join will screw up scheme URIs
	// (20170124/thisisaaronland)

	_, err = url.Parse(root)

	if err == nil {

		if !strings.HasSuffix(root, "/") {
			root += "/"
		}

		abs_path = root + rel

	} else {
		abs_path = filepath.Join(root, rel)
	}

	return abs_path, nil
}

func IsWOFFile(path string) (bool, error) {

	re_woffile, err := regexp.Compile(`^\d+(?:\-alt\-.*)?\.geojson$`)

	if err != nil {
		return false, err
	}

	abs_path, err := filepath.Abs(path)

	if err != nil {
		return false, err
	}

	fname := filepath.Base(abs_path)

	wof := re_woffile.MatchString(fname)

	return wof, nil
}

func IsAltFile(path string) (bool, error) {

	alt, err := AltGeomFromPath(path)

	if err != nil {
		return false, err
	}

	if alt == nil {
		return false, nil
	}

	return true, nil
}

func AltGeomFromPath(path string) (*AltGeom, error) {

	re_altfile, err := regexp.Compile(`^\d+\-alt\-(.*)\.geojson$`)

	if err != nil {
		return nil, err
	}

	abs_path, err := filepath.Abs(path)

	if err != nil {
		return nil, err
	}

	fname := filepath.Base(abs_path)

	m := re_altfile.FindStringSubmatch(fname)

	if len(m) == 0 {
		return nil, nil
	}

	str_parts := m[1]
	parts := strings.Split(str_parts, "-")

	alt := AltGeom{
		Source: parts[0],
	}

	if len(parts) >= 2 {
		alt.Function = parts[1]
	}

	if len(parts) >= 3 {
		alt.Extras = parts[2:]
	}

	return &alt, nil
}

func IdFromPath(path string) (int64, error) {

	abs_path, err := filepath.Abs(path)

	if err != nil {
		return -1, err
	}

	ok, err := IsWOFFile(abs_path)

	if err != nil {
		return -1, err
	}

	if !ok {
		return -1, errors.New("Not a valid WOF file")
	}

	fname := filepath.Base(abs_path)

	re_wofid, err := regexp.Compile(`^(\d+)(?:\-alt\-.*)?\.geojson$`)

	if err != nil {
		return -1, err
	}

	match := re_wofid.FindAllStringSubmatch(fname, -1)

	if len(match[0]) != 2 {
		return -1, errors.New("Unable to parse filename")
	}

	wofid, err := strconv.ParseInt(match[0][1], 10, 64)

	if err != nil {
		return -1, err
	}

	return wofid, nil
}

func RepoFromPath(path string) (string, error) {

	abs_path, err := filepath.Abs(path)

	if err != nil {
		return "", err
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
		return "", errors.New("Unable to determine repo from path")
	}

	return repo, nil
}
