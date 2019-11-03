package feature

import (
	"encoding/json"
	"fmt"
	"github.com/skelterjohn/geom"
	"github.com/whosonfirst/go-whosonfirst-flags"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/geometry"
	props_wof "github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	"github.com/whosonfirst/go-whosonfirst-spr"
	"github.com/whosonfirst/warning"
	"strconv"
)

type WOFAltFeature struct {
	geojson.Feature
	body []byte
}

type WOFAltStandardPlacesResult struct {
	spr.StandardPlacesResult `json:",omitempty"`
	SPRId                    string  `json:"spr:id"`
	SPRName                  string  `json:"spr:name"`
	SPRPlacetype             string  `json:"spr:placetype"`
	SPRLatitude              float64 `json:"spr:latitude"`
	SPRLongitude             float64 `json:"spr:longitude"`
	SPRMinLatitude           float64 `json:"spr:min_latitude"`
	SPRMinLongitude          float64 `json:"spr:min_longitude"`
	SPRMaxLatitude           float64 `json:"spr:max_latitude"`
	SPRMaxLongitude          float64 `json:"spr:max_longitude"`
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

	possible := []string{
		"properties.src:geom",
	}

	src_geom := utils.StringProperty(f.Bytes(), possible, "unknown")

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

	bboxes, err := f.BoundingBoxes()

	if err != nil {
		return nil, err
	}

	mbr := bboxes.MBR()

	lat := mbr.Min.Y + ((mbr.Max.Y - mbr.Min.Y) / 2.0)
	lon := mbr.Min.X + ((mbr.Max.X - mbr.Min.X) / 2.0)

	spr := WOFAltStandardPlacesResult{
		SPRId:           f.Id(),
		SPRPlacetype:    f.Placetype(),
		SPRName:         f.Name(),
		SPRLatitude:     lat,
		SPRLongitude:    lon,
		SPRMinLatitude:  mbr.Min.Y,
		SPRMinLongitude: mbr.Min.X,
		SPRMaxLatitude:  mbr.Max.Y,
		SPRMaxLongitude: mbr.Max.X,
	}

	return &spr, nil
}

func (spr *WOFAltStandardPlacesResult) Id() string {
	return spr.SPRId
}

func (spr *WOFAltStandardPlacesResult) ParentId() string {
	return ""
}

func (spr *WOFAltStandardPlacesResult) Name() string {
	return spr.SPRName
}

func (spr *WOFAltStandardPlacesResult) Placetype() string {
	return spr.SPRPlacetype
}

func (spr *WOFAltStandardPlacesResult) Country() string {
	return "XX"
}

func (spr *WOFAltStandardPlacesResult) Repo() string {
	return ""
}

func (spr *WOFAltStandardPlacesResult) Path() string {
	return ""
}

func (spr *WOFAltStandardPlacesResult) URI() string {
	return ""
}

func (spr *WOFAltStandardPlacesResult) Latitude() float64 {
	return spr.SPRLatitude
}

func (spr *WOFAltStandardPlacesResult) Longitude() float64 {
	return spr.SPRLongitude
}

func (spr *WOFAltStandardPlacesResult) MinLatitude() float64 {
	return spr.SPRMinLatitude
}

func (spr *WOFAltStandardPlacesResult) MinLongitude() float64 {
	return spr.SPRMinLongitude
}

func (spr *WOFAltStandardPlacesResult) MaxLatitude() float64 {
	return spr.SPRLatitude
}

func (spr *WOFAltStandardPlacesResult) MaxLongitude() float64 {
	return spr.SPRMaxLongitude
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

func (spr *WOFAltStandardPlacesResult) LastModified() int64 {
	return -1
}
