package feature

import (
	"encoding/json"
	"github.com/sfomuseum/go-edtf"
	"github.com/skelterjohn/geom"
	"github.com/whosonfirst/go-whosonfirst-flags"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/geometry"
	props_geom "github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/geometry"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"strings"
)

type GeoJSONFeature struct {
	geojson.Feature
	body []byte
}

type GeoJSONStandardPlacesResult struct {
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

func NewGeoJSONFeature(body []byte) (geojson.Feature, error) {

	var stub interface{}
	err := json.Unmarshal(body, &stub)

	if err != nil {
		return nil, err
	}

	f := GeoJSONFeature{
		body: body,
	}

	return &f, nil
}

func (f *GeoJSONFeature) ContainsCoord(c geom.Coord) (bool, error) {

	return geometry.FeatureContainsCoord(f, c)
}

func (f *GeoJSONFeature) String() string {

	body, err := json.Marshal(f.body)

	if err != nil {
		return ""
	}

	return string(body)
}

func (f *GeoJSONFeature) Bytes() []byte {

	return f.body
}

func (f *GeoJSONFeature) Id() string {

	possible := []string{
		"id",
		"properties.id",
	}

	id := utils.StringProperty(f.Bytes(), possible, "")

	if id == "" {
		id = f.uid()
	}

	return id
}

func (f *GeoJSONFeature) Name() string {

	possible := []string{
		"properties.name",
	}

	name := utils.StringProperty(f.Bytes(), possible, "")

	if name == "" {
		name = f.uid()
	}

	return name
}

func (f *GeoJSONFeature) Placetype() string {

	possible := []string{
		"properties.placetype",
	}

	pt := utils.StringProperty(f.Bytes(), possible, "")

	if pt == "" {
		pt = props_geom.Type(f)
		pt = strings.ToLower(pt)
	}

	return pt
}

func (f *GeoJSONFeature) uid() string {

	h, err := utils.GeohashFeature(f)

	if err != nil {
		h = "..."
	}

	return h
}

func (f *GeoJSONFeature) BoundingBoxes() (geojson.BoundingBoxes, error) {
	return geometry.BoundingBoxesForFeature(f)
}

func (f *GeoJSONFeature) Polygons() ([]geojson.Polygon, error) {
	return geometry.PolygonsForFeature(f)
}

func (f *GeoJSONFeature) SPR() (spr.StandardPlacesResult, error) {

	bboxes, err := f.BoundingBoxes()

	if err != nil {
		return nil, err
	}

	mbr := bboxes.MBR()

	lat := mbr.Min.Y + ((mbr.Max.Y - mbr.Min.Y) / 2.0)
	lon := mbr.Min.X + ((mbr.Max.X - mbr.Min.X) / 2.0)

	spr := GeoJSONStandardPlacesResult{
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

func (spr *GeoJSONStandardPlacesResult) Id() string {
	return spr.SPRId
}

func (spr *GeoJSONStandardPlacesResult) ParentId() string {
	return ""
}

func (spr *GeoJSONStandardPlacesResult) Name() string {
	return spr.SPRName
}

func (spr *GeoJSONStandardPlacesResult) Placetype() string {
	return spr.SPRPlacetype
}

func (spr *GeoJSONStandardPlacesResult) Inception() *edtf.EDTFDate {
	return nil
}

func (spr *GeoJSONStandardPlacesResult) Cessation() *edtf.EDTFDate {
	return nil
}

func (spr *GeoJSONStandardPlacesResult) Country() string {
	return "XX"
}

func (spr *GeoJSONStandardPlacesResult) Repo() string {
	return ""
}

func (spr *GeoJSONStandardPlacesResult) Path() string {
	return ""
}

func (spr *GeoJSONStandardPlacesResult) URI() string {
	return ""
}

func (spr *GeoJSONStandardPlacesResult) Latitude() float64 {
	return spr.SPRLatitude
}

func (spr *GeoJSONStandardPlacesResult) Longitude() float64 {
	return spr.SPRLongitude
}

func (spr *GeoJSONStandardPlacesResult) MinLatitude() float64 {
	return spr.SPRMinLatitude
}

func (spr *GeoJSONStandardPlacesResult) MinLongitude() float64 {
	return spr.SPRMinLongitude
}

func (spr *GeoJSONStandardPlacesResult) MaxLatitude() float64 {
	return spr.SPRLatitude
}

func (spr *GeoJSONStandardPlacesResult) MaxLongitude() float64 {
	return spr.SPRMaxLongitude
}

func (spr *GeoJSONStandardPlacesResult) IsCurrent() flags.ExistentialFlag {
	return existentialFlag(-1)
}

func (spr *GeoJSONStandardPlacesResult) IsCeased() flags.ExistentialFlag {
	return existentialFlag(-1)
}

func (spr *GeoJSONStandardPlacesResult) IsDeprecated() flags.ExistentialFlag {
	return existentialFlag(-1)
}

func (spr *GeoJSONStandardPlacesResult) IsSuperseded() flags.ExistentialFlag {
	return existentialFlag(-1)
}

func (spr *GeoJSONStandardPlacesResult) IsSuperseding() flags.ExistentialFlag {
	return existentialFlag(-1)
}

func (spr *GeoJSONStandardPlacesResult) SupersededBy() []int64 {
	return []int64{}
}

func (spr *GeoJSONStandardPlacesResult) Supersedes() []int64 {
	return []int64{}
}

func (spr *GeoJSONStandardPlacesResult) BelongsTo() []int64 {
	return []int64{}
}

func (spr *GeoJSONStandardPlacesResult) LastModified() int64 {
	return -1
}
