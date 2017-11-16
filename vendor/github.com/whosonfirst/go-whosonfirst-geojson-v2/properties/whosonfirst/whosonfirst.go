package whosonfirst

import (
	"encoding/json"
	"github.com/skelterjohn/geom"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-flags"
	"github.com/whosonfirst/go-whosonfirst-flags/existential"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	"strings"
)

type WOFConcordances map[string]string

type WOFCentroid struct {
	geojson.Centroid
	coord  geom.Coord
	source string
}

func (c *WOFCentroid) Coord() geom.Coord {
	return c.coord
}

func (c *WOFCentroid) Source() string {
	return c.source
}

func (c *WOFCentroid) ToString() (string, error) {

	type Geometry struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	}

	g := Geometry{
		Type:        "Point",
		Coordinates: []float64{c.coord.X, c.coord.Y},
	}

	b, err := json.Marshal(g)

	if err != nil {
		return "", err
	}

	return string(b), nil
}

func NewWOFCentroid(lat float64, lon float64, source string) (geojson.Centroid, error) {

	coord, err := utils.NewCoordinateFromLatLons(lat, lon)

	if err != nil {
		return nil, err
	}

	c := WOFCentroid{
		coord:  coord,
		source: source,
	}

	return &c, nil
}

func Centroid(f geojson.Feature) (geojson.Centroid, error) {

	var lat gjson.Result
	var lon gjson.Result

	lat = gjson.GetBytes(f.Bytes(), "properties.lbl:latitude")
	lon = gjson.GetBytes(f.Bytes(), "properties.lbl:longitude")

	if lat.Exists() && lon.Exists() {
		return NewWOFCentroid(lat.Float(), lon.Float(), "lbl")
	}

	lat = gjson.GetBytes(f.Bytes(), "properties.reversegeo:latitude")
	lon = gjson.GetBytes(f.Bytes(), "properties.reversegeo:longitude")

	if lat.Exists() && lon.Exists() {
		return NewWOFCentroid(lat.Float(), lon.Float(), "reversegeo")
	}

	lat = gjson.GetBytes(f.Bytes(), "properties.geom:latitude")
	lon = gjson.GetBytes(f.Bytes(), "properties.geom:longitude")

	if lat.Exists() && lon.Exists() {
		return NewWOFCentroid(lat.Float(), lon.Float(), "geom")
	}

	return NewWOFCentroid(0.0, 0.0, "nullisland")
}

func Concordances(f geojson.Feature) (WOFConcordances, error) {

	concordances := make(map[string]string)

	rsp := gjson.GetBytes(f.Bytes(), "properties.wof:concordances")

	if !rsp.Exists() {
		return concordances, nil
	}

	for k, v := range rsp.Map() {
		concordances[k] = v.String()
	}

	return concordances, nil
}

func Country(f geojson.Feature) string {

	possible := []string{
		"properties.wof:country",
	}

	return utils.StringProperty(f.Bytes(), possible, "XX")
}

func Id(f geojson.Feature) int64 {

	possible := []string{
		"properties.wof:id",
		"id",
	}

	return utils.Int64Property(f.Bytes(), possible, -1)
}

func Name(f geojson.Feature) string {

	possible := []string{
		"properties.wof:name",
		"properties.name",
	}

	return utils.StringProperty(f.Bytes(), possible, "a place with no name")
}

func ParentId(f geojson.Feature) int64 {

	possible := []string{
		"properties.wof:parent_id",
	}

	return utils.Int64Property(f.Bytes(), possible, -1)
}

func Placetype(f geojson.Feature) string {

	possible := []string{
		"properties.wof:placetype",
		"properties.placetype",
	}

	return utils.StringProperty(f.Bytes(), possible, "here be dragons")
}

func Repo(f geojson.Feature) string {

	possible := []string{
		"properties.wof:repo",
	}

	return utils.StringProperty(f.Bytes(), possible, "whosonfirst-data-xx")
}

func LastModified(f geojson.Feature) int64 {

	possible := []string{
		"properties.wof:lastmodified",
	}

	return utils.Int64Property(f.Bytes(), possible, -1)
}

func IsCurrent(f geojson.Feature) (flags.ExistentialFlag, error) {

	possible := []string{
		"properties.mz:is_current",
	}

	v := utils.Int64Property(f.Bytes(), possible, -1)

	if v == 1 || v == 0 {
		return existential.NewKnownUnknownFlag(v)
	}

	d, err := IsDeprecated(f)

	if err != nil {
		return nil, err
	}

	if d.IsTrue() && d.IsKnown() {
		return existential.NewKnownUnknownFlag(0)
	}

	c, err := IsCeased(f)

	if err != nil {
		return nil, err
	}

	if c.IsTrue() && c.IsKnown() {
		return existential.NewKnownUnknownFlag(0)
	}

	s, err := IsSuperseded(f)

	if err != nil {
		return nil, err
	}

	if s.IsTrue() && s.IsKnown() {
		return existential.NewKnownUnknownFlag(0)
	}

	return existential.NewKnownUnknownFlag(-1)
}

func IsDeprecated(f geojson.Feature) (flags.ExistentialFlag, error) {

	possible := []string{
		"properties.edtf:deprecated",
	}

	v := utils.StringProperty(f.Bytes(), possible, "-")

	switch v {
	case "-":
		return existential.NewKnownUnknownFlag(0)
	case "u":
		return existential.NewKnownUnknownFlag(-1)
	case "uuuu":
		return existential.NewKnownUnknownFlag(-1)
	default:
		return existential.NewKnownUnknownFlag(1)
	}
}

func IsCeased(f geojson.Feature) (flags.ExistentialFlag, error) {

	possible := []string{
		"properties.edtf:cessation",
	}

	v := utils.StringProperty(f.Bytes(), possible, "uuuu")

	switch v {
	case "":
		return existential.NewKnownUnknownFlag(0)
	case "u":
		return existential.NewKnownUnknownFlag(-1)
	case "uuuu":
		return existential.NewKnownUnknownFlag(-1)
	default:
		return existential.NewKnownUnknownFlag(1)
	}
}

func IsSuperseded(f geojson.Feature) (flags.ExistentialFlag, error) {

	by := gjson.GetBytes(f.Bytes(), "properties.wof:superseded_by")

	if by.Exists() && len(by.Array()) > 0 {
		return existential.NewKnownUnknownFlag(1)
	}

	return existential.NewKnownUnknownFlag(0)
}

func IsSuperseding(f geojson.Feature) (flags.ExistentialFlag, error) {

	sc := gjson.GetBytes(f.Bytes(), "properties.wof:supersedes")

	if sc.Exists() && len(sc.Array()) > 0 {
		return existential.NewKnownUnknownFlag(1)
	}

	return existential.NewKnownUnknownFlag(0)
}

func SupersededBy(f geojson.Feature) []int64 {

	superseded_by := make([]int64, 0)

	possible := gjson.GetBytes(f.Bytes(), "properties.wof:superseded_by")

	if possible.Exists() {

		for _, id := range possible.Array() {
			superseded_by = append(superseded_by, id.Int())
		}
	}

	return superseded_by
}

func Supersedes(f geojson.Feature) []int64 {

	supersedes := make([]int64, 0)

	possible := gjson.GetBytes(f.Bytes(), "properties.wof:supersedes")

	if possible.Exists() {

		for _, id := range possible.Array() {
			supersedes = append(supersedes, id.Int())
		}
	}

	return supersedes
}

func Names(f geojson.Feature) map[string][]string {

	names_map := make(map[string][]string)

	r := gjson.GetBytes(f.Bytes(), "properties")

	if !r.Exists() {
		return names_map
	}

	for k, v := range r.Map() {

		if !strings.HasPrefix(k, "name:") {
			continue
		}

		if !v.Exists() {
			continue
		}

		name := strings.Replace(k, "name:", "", 1)
		names := make([]string, 0)

		for _, n := range v.Array() {
			names = append(names, n.String())
		}

		names_map[name] = names
	}

	return names_map
}

// DEPRECATED - PLEASE FOR TO BE USING Hierarchies

func Hierarchy(f geojson.Feature) []map[string]int64 {
	return Hierarchies(f)
}

func Hierarchies(f geojson.Feature) []map[string]int64 {

	hierarchies := make([]map[string]int64, 0)

	possible := gjson.GetBytes(f.Bytes(), "properties.wof:hierarchy")

	if possible.Exists() {

		for _, h := range possible.Array() {

			foo := make(map[string]int64)

			for k, v := range h.Map() {

				foo[k] = v.Int()
			}

			hierarchies = append(hierarchies, foo)
		}
	}

	return hierarchies
}
