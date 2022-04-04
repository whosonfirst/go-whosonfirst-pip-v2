package feature

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sfomuseum/go-edtf"
	"github.com/skelterjohn/geom"
	"github.com/whosonfirst/go-whosonfirst-flags"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/geometry"
	props_wof "github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/warning"
	"strconv"
	"strings"
)

type WOFAltFeature struct {
	geojson.Feature
	body []byte
}

type WOFAltStandardPlacesResult struct {
	spr.StandardPlacesResult `json:",omitempty"`
	WOFId                    string  `json:"wof:id"`
	WOFName                  string  `json:"wof:name"`
	WOFPlacetype             string  `json:"wof:placetype"`
	MZLatitude               float64 `json:"mz:latitude"`
	MZLongitude              float64 `json:"mz:longitude"`
	MZMinLatitude            float64 `json:"mz:min_latitude"`
	MZMinLongitude           float64 `json:"mz:min_longitude"`
	MZMaxLatitude            float64 `json:"mz:max_latitude"`
	MZMaxLongitude           float64 `json:"mz:max_longitude"`
	WOFPath                  string  `json:"wof:path"`
	WOFRepo                  string  `json:"wof:repo"`
}

func EnsureWOFAltFeature(body []byte) error {

	required := []string{
		"properties.wof:id",
		"properties.wof:repo",
		"properties.src:alt_label",
	}

	err := utils.EnsureProperties(body, required)

	if err != nil {
		return err
	}

	return nil
}

func NewWOFAltFeature(body []byte) (geojson.Feature, error) {

	var stub interface{}
	err := json.Unmarshal(body, &stub)

	if err != nil {
		return nil, err
	}

	err = EnsureWOFAltFeature(body)

	if err != nil && !warning.IsWarning(err) {
		return nil, err
	}

	f := WOFAltFeature{
		body: body,
	}

	return &f, nil
}

func (f *WOFAltFeature) ContainsCoord(c geom.Coord) (bool, error) {

	return geometry.FeatureContainsCoord(f, c)
}

func (f *WOFAltFeature) String() string {

	body, err := json.Marshal(f.body)

	if err != nil {
		return ""
	}

	return string(body)
}

func (f *WOFAltFeature) Bytes() []byte {

	return f.body
}

func (f *WOFAltFeature) Id() string {

	id := props_wof.Id(f)
	return strconv.FormatInt(id, 10)
}

func (f *WOFAltFeature) Name() string {

	id := f.Id()

	src_geom := props_wof.Source(f)

	return fmt.Sprintf("%s alt geometry (%s)", id, src_geom)
}

func (f *WOFAltFeature) Placetype() string {
	return "alt"
}

func (f *WOFAltFeature) BoundingBoxes() (geojson.BoundingBoxes, error) {
	return geometry.BoundingBoxesForFeature(f)
}

func (f *WOFAltFeature) Polygons() ([]geojson.Polygon, error) {
	return geometry.PolygonsForFeature(f)
}

func (f *WOFAltFeature) SPR() (spr.StandardPlacesResult, error) {

	id := props_wof.Id(f)
	alt_label := props_wof.AltLabel(f)
	label_parts := strings.Split(alt_label, "-")

	if len(label_parts) == 0 {
		return nil, errors.New("Invalid src:alt_label property")
	}

	alt_geom := &uri.AltGeom{
		Source: label_parts[0],
	}

	if len(label_parts) >= 2 {
		alt_geom.Function = label_parts[1]
	}

	if len(label_parts) >= 3 {
		alt_geom.Extras = label_parts[2:]
	}

	uri_args := &uri.URIArgs{
		IsAlternate: true,
		AltGeom:     alt_geom,
	}

	rel_path, err := uri.Id2RelPath(id, uri_args)

	if err != nil {
		return nil, err
	}

	repo := props_wof.Repo(f)

	bboxes, err := f.BoundingBoxes()

	if err != nil {
		return nil, err
	}

	mbr := bboxes.MBR()

	lat := mbr.Min.Y + ((mbr.Max.Y - mbr.Min.Y) / 2.0)
	lon := mbr.Min.X + ((mbr.Max.X - mbr.Min.X) / 2.0)

	spr := WOFAltStandardPlacesResult{
		WOFId:          f.Id(),
		WOFPlacetype:   f.Placetype(),
		WOFName:        f.Name(),
		MZLatitude:     lat,
		MZLongitude:    lon,
		MZMinLatitude:  mbr.Min.Y,
		MZMinLongitude: mbr.Min.X,
		MZMaxLatitude:  mbr.Max.Y,
		MZMaxLongitude: mbr.Max.X,
		WOFPath:        rel_path,
		WOFRepo:        repo,
	}

	return &spr, nil
}

func (spr *WOFAltStandardPlacesResult) Id() string {
	return spr.WOFId
}

func (spr *WOFAltStandardPlacesResult) ParentId() string {
	return "-1"
}

func (spr *WOFAltStandardPlacesResult) Name() string {
	return spr.WOFName
}

func (spr *WOFAltStandardPlacesResult) Placetype() string {
	return spr.WOFPlacetype
}

func (spr *WOFAltStandardPlacesResult) Country() string {
	return "XX"
}

func (spr *WOFAltStandardPlacesResult) Repo() string {
	return spr.WOFRepo
}

func (spr *WOFAltStandardPlacesResult) Path() string {
	return spr.WOFPath
}

func (spr *WOFAltStandardPlacesResult) URI() string {
	return ""
}

func (spr *WOFAltStandardPlacesResult) Latitude() float64 {
	return spr.MZLatitude
}

func (spr *WOFAltStandardPlacesResult) Longitude() float64 {
	return spr.MZLongitude
}

func (spr *WOFAltStandardPlacesResult) MinLatitude() float64 {
	return spr.MZMinLatitude
}

func (spr *WOFAltStandardPlacesResult) MinLongitude() float64 {
	return spr.MZMinLongitude
}

func (spr *WOFAltStandardPlacesResult) MaxLatitude() float64 {
	return spr.MZLatitude
}

func (spr *WOFAltStandardPlacesResult) MaxLongitude() float64 {
	return spr.MZMaxLongitude
}

func (spr *WOFAltStandardPlacesResult) Inception() *edtf.EDTFDate {
	return nil
}

func (spr *WOFAltStandardPlacesResult) Cessation() *edtf.EDTFDate {
	return nil
}

func (spr *WOFAltStandardPlacesResult) IsCurrent() flags.ExistentialFlag {
	return existentialFlag(-1)
}

func (spr *WOFAltStandardPlacesResult) IsCeased() flags.ExistentialFlag {
	return existentialFlag(-1)
}

func (spr *WOFAltStandardPlacesResult) IsDeprecated() flags.ExistentialFlag {
	return existentialFlag(-1)
}

func (spr *WOFAltStandardPlacesResult) IsSuperseded() flags.ExistentialFlag {
	return existentialFlag(-1)
}

func (spr *WOFAltStandardPlacesResult) IsSuperseding() flags.ExistentialFlag {
	return existentialFlag(-1)
}

func (spr *WOFAltStandardPlacesResult) SupersededBy() []int64 {
	return []int64{}
}

func (spr *WOFAltStandardPlacesResult) Supersedes() []int64 {
	return []int64{}
}

func (spr *WOFAltStandardPlacesResult) BelongsTo() []int64 {
	return []int64{}
}

func (spr *WOFAltStandardPlacesResult) LastModified() int64 {
	return -1
}
