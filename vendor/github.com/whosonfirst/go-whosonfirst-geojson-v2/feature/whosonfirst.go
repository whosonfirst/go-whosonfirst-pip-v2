package feature

import (
	"encoding/json"
	_ "errors"
	"github.com/sfomuseum/go-edtf"
	"github.com/sfomuseum/go-edtf/parser"
	"github.com/skelterjohn/geom"
	"github.com/whosonfirst/go-whosonfirst-flags"
	"github.com/whosonfirst/go-whosonfirst-flags/existential"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/geometry"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	"github.com/whosonfirst/go-whosonfirst-placetypes"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/warning"
	"strconv"
)

type WOFFeature struct {
	geojson.Feature
	body []byte
}

type WOFStandardPlacesResult struct {
	spr.StandardPlacesResult `json:",omitempty"`
	EDTFInception            string  `json:"edtf:inception"`
	EDTFCessation            string  `json:"edtf:cessation"`
	WOFId                    int64   `json:"wof:id"`
	WOFParentId              int64   `json:"wof:parent_id"`
	WOFName                  string  `json:"wof:name"`
	WOFPlacetype             string  `json:"wof:placetype"`
	WOFCountry               string  `json:"wof:country"`
	WOFRepo                  string  `json:"wof:repo"`
	WOFPath                  string  `json:"wof:path"`
	WOFSupersededBy          []int64 `json:"wof:superseded_by"`
	WOFSupersedes            []int64 `json:"wof:supersedes"`
	WOFBelongsTo             []int64 `json:"wof:belongsto"`
	MZURI                    string  `json:"mz:uri"`
	MZLatitude               float64 `json:"mz:latitude"`
	MZLongitude              float64 `json:"mz:longitude"`
	MZMinLatitude            float64 `json:"mz:min_latitude"`
	MZMinLongitude           float64 `json:"mz:min_longitude"`
	MZMaxLatitude            float64 `json:"mz:max_latitude"`
	MZMaxLongitude           float64 `json:"mz:max_longitude"`
	MZIsCurrent              int64   `json:"mz:is_current"`
	MZIsCeased               int64   `json:"mz:is_ceased"`
	MZIsDeprecated           int64   `json:"mz:is_deprecated"`
	MZIsSuperseded           int64   `json:"mz:is_superseded"`
	MZIsSuperseding          int64   `json:"mz:is_superseding"`
	WOFLastModified          int64   `json:"wof:lastmodified"`
}

func EnsureWOFFeature(body []byte) error {

	required := []string{
		"properties.wof:id",
		"properties.wof:name",
		"properties.wof:repo",
		"properties.wof:placetype",
		// we used to handle these like this but we
		// do some jiggling below to account for the
		// fact that we might working with an SPR...
		// "properties.geom:latitude",
		// "properties.geom:longitude",
		// "properties.geom:bbox",
	}

	err := utils.EnsureProperties(body, required)

	if err != nil {
		return err
	}

	// strictly speaking we probably want to ensure all if the spr_geom
	// properties if we have to test one of them but let's see how this
	// works first... (20180223/thisisaaronland)

	required_geom := map[string][]string{
		"properties.geom:latitude":  []string{"properties.mz:latitude"},
		"properties.geom:longitude": []string{"properties.mz:latitude"},
		"properties.geom:bbox":      []string{"properties.mz:min_latitude", "properties.mz:min_longitude", "properties.mz:max_latitude", "properties.mz:max_longitude"},
	}

	for wof_geom, spr_geom := range required_geom {

		err = utils.EnsureProperties(body, []string{wof_geom})

		if err == nil {
			continue
		}

		err = utils.EnsureProperties(body, spr_geom)

		if err != nil {
			return err
		}
	}

	// we may want or need to handle WOF documents with placetypes
	// not already defined in core (like for anyone working on datasets
	// outside the scope of core...) / there is an open branch of the
	// go-whosonfirst-placetypes package for adding custom placetypes
	// but it's not at all clear whose vendor-ed (go-wof-pt) package
	// will get used so never mind that / we could also add a global flag
	// to this package to disable checks but on measure it seems best
	// to issue a warning thing that implements the error interface and
	// leave the details to individual applications / we are using a
	// forked (to the whosonfirst org) version of https://github.com/lunemec/warning
	// (20180405/thisisaaronland)

	pt := utils.StringProperty(body, []string{"properties.wof:placetype"}, "")

	if !placetypes.IsValidPlacetype(pt) {
		return warning.New("Invalid wof:placetype")
	}

	// check wof:repo here?

	return nil
}

func NewWOFFeature(body []byte) (geojson.Feature, error) {

	var stub interface{}
	err := json.Unmarshal(body, &stub)

	if err != nil {
		return nil, err
	}

	err = EnsureWOFFeature(body)

	if err != nil && !warning.IsWarning(err) {
		return nil, err
	}

	f := WOFFeature{
		body: body,
	}

	// because err might be a warning.Error / see notes above in EnsureWOFFeature
	// I don't really love this... (20180405/thisisaaronland)

	return &f, err
}

func (f *WOFFeature) String() string {

	body, err := json.Marshal(f.body)

	if err != nil {
		return ""
	}

	return string(body)
}

func (f *WOFFeature) Bytes() []byte {
	return f.body
}

func (f *WOFFeature) Id() string {
	id := whosonfirst.Id(f)
	return strconv.FormatInt(id, 10)
}

func (f *WOFFeature) Name() string {
	return whosonfirst.Name(f)
}

func (f *WOFFeature) Placetype() string {
	return whosonfirst.Placetype(f)
}

func (f *WOFFeature) BoundingBoxes() (geojson.BoundingBoxes, error) {
	return geometry.BoundingBoxesForFeature(f)
}

func (f *WOFFeature) Polygons() ([]geojson.Polygon, error) {
	return geometry.PolygonsForFeature(f)
}

func (f *WOFFeature) ContainsCoord(c geom.Coord) (bool, error) {
	return geometry.FeatureContainsCoord(f, c)
}

func (f *WOFFeature) SPR() (spr.StandardPlacesResult, error) {

	id := whosonfirst.Id(f)
	parent_id := whosonfirst.ParentId(f)
	name := whosonfirst.Name(f)
	placetype := whosonfirst.Placetype(f)
	country := whosonfirst.Country(f)
	repo := whosonfirst.Repo(f)

	inception := whosonfirst.Inception(f)
	cessation := whosonfirst.Cessation(f)

	// See this: We're accounting for all the pre-2019 EDTF spec
	// inception but mostly cessation strings by silently swapping
	// them out (20210321/straup)

	_, err := parser.ParseString(inception)

	if err != nil {

		if !isDeprecatedEDTF(inception) {
			return nil, err
		}

		replacement, err := replaceDeprecatedEDTF(inception)

		if err != nil {
			return nil, err
		}

		inception = replacement
	}

	_, err = parser.ParseString(cessation)

	if err != nil {

		if !isDeprecatedEDTF(cessation) {
			return nil, err
		}

		replacement, err := replaceDeprecatedEDTF(cessation)

		if err != nil {
			return nil, err
		}

		cessation = replacement
	}

	path, err := uri.Id2RelPath(id)

	if err != nil {
		return nil, err
	}

	uri, err := uri.Id2AbsPath("https://data.whosonfirst.org", id)

	if err != nil {
		return nil, err
	}

	is_current, err := whosonfirst.IsCurrent(f)

	if err != nil {
		return nil, err
	}

	is_ceased, err := whosonfirst.IsCeased(f)

	if err != nil {
		return nil, err
	}

	is_deprecated, err := whosonfirst.IsDeprecated(f)

	if err != nil {
		return nil, err
	}

	is_superseded, err := whosonfirst.IsSuperseded(f)

	if err != nil {
		return nil, err
	}

	is_superseding, err := whosonfirst.IsSuperseding(f)

	if err != nil {
		return nil, err
	}

	centroid, err := whosonfirst.Centroid(f)

	if err != nil {
		return nil, err
	}

	bboxes, err := f.BoundingBoxes()

	if err != nil {
		return nil, err
	}

	coord := centroid.Coord()
	mbr := bboxes.MBR()

	superseded_by := whosonfirst.SupersededBy(f)
	supersedes := whosonfirst.Supersedes(f)
	belongsto := whosonfirst.BelongsTo(f)

	lastmod := whosonfirst.LastModified(f)

	spr := WOFStandardPlacesResult{
		WOFId:           id,
		WOFParentId:     parent_id,
		WOFPlacetype:    placetype,
		WOFName:         name,
		WOFCountry:      country,
		WOFRepo:         repo,
		WOFPath:         path,
		WOFSupersedes:   supersedes,
		WOFSupersededBy: superseded_by,
		WOFBelongsTo:    belongsto,
		EDTFInception:   inception,
		EDTFCessation:   cessation,
		MZURI:           uri,
		MZLatitude:      coord.Y,
		MZLongitude:     coord.X,
		MZMinLatitude:   mbr.Min.Y,
		MZMinLongitude:  mbr.Min.X,
		MZMaxLatitude:   mbr.Max.Y,
		MZMaxLongitude:  mbr.Max.X,
		MZIsCurrent:     is_current.Flag(),
		MZIsCeased:      is_ceased.Flag(),
		MZIsDeprecated:  is_deprecated.Flag(),
		MZIsSuperseded:  is_superseded.Flag(),
		MZIsSuperseding: is_superseding.Flag(),
		WOFLastModified: lastmod,
	}

	return &spr, nil
}

func (spr *WOFStandardPlacesResult) Id() string {
	return strconv.FormatInt(spr.WOFId, 10)
}

func (spr *WOFStandardPlacesResult) ParentId() string {
	return strconv.FormatInt(spr.WOFParentId, 10)
}

func (spr *WOFStandardPlacesResult) Name() string {
	return spr.WOFName
}

func (spr *WOFStandardPlacesResult) Inception() *edtf.EDTFDate {
	return spr.edtfDate(spr.EDTFInception)
}

func (spr *WOFStandardPlacesResult) Cessation() *edtf.EDTFDate {
	return spr.edtfDate(spr.EDTFCessation)
}

func (spr *WOFStandardPlacesResult) edtfDate(edtf_str string) *edtf.EDTFDate {

	d, err := parser.ParseString(edtf_str)

	if err != nil {
		return nil
	}

	return d
}

func (spr *WOFStandardPlacesResult) Placetype() string {
	return spr.WOFPlacetype
}

func (spr *WOFStandardPlacesResult) Country() string {
	return spr.WOFCountry
}

func (spr *WOFStandardPlacesResult) Repo() string {
	return spr.WOFRepo
}

func (spr *WOFStandardPlacesResult) Path() string {
	return spr.WOFPath
}

func (spr *WOFStandardPlacesResult) URI() string {
	return spr.MZURI
}

func (spr *WOFStandardPlacesResult) Latitude() float64 {
	return spr.MZLatitude
}

func (spr *WOFStandardPlacesResult) Longitude() float64 {
	return spr.MZLongitude
}

func (spr *WOFStandardPlacesResult) MinLatitude() float64 {
	return spr.MZMinLatitude
}

func (spr *WOFStandardPlacesResult) MinLongitude() float64 {
	return spr.MZMinLongitude
}

func (spr *WOFStandardPlacesResult) MaxLatitude() float64 {
	return spr.MZLatitude
}

func (spr *WOFStandardPlacesResult) MaxLongitude() float64 {
	return spr.MZMaxLongitude
}

func (spr *WOFStandardPlacesResult) IsCurrent() flags.ExistentialFlag {
	return existentialFlag(spr.MZIsCurrent)
}

func (spr *WOFStandardPlacesResult) IsCeased() flags.ExistentialFlag {
	return existentialFlag(spr.MZIsCeased)
}

func (spr *WOFStandardPlacesResult) IsDeprecated() flags.ExistentialFlag {
	return existentialFlag(spr.MZIsDeprecated)
}

func (spr *WOFStandardPlacesResult) IsSuperseded() flags.ExistentialFlag {
	return existentialFlag(spr.MZIsSuperseded)
}

func (spr *WOFStandardPlacesResult) IsSuperseding() flags.ExistentialFlag {
	return existentialFlag(spr.MZIsSuperseding)
}

func (spr *WOFStandardPlacesResult) SupersededBy() []int64 {
	return spr.WOFSupersededBy
}

func (spr *WOFStandardPlacesResult) Supersedes() []int64 {
	return spr.WOFSupersedes
}

func (spr *WOFStandardPlacesResult) BelongsTo() []int64 {
	return spr.WOFBelongsTo
}

func (spr *WOFStandardPlacesResult) LastModified() int64 {
	return spr.WOFLastModified
}

// we're going to assume that this won't fail since we already go through
// the process of instantiating `flags.ExistentialFlag` thingies in SPR()
// if we need to we'll just cache those instances in the `spr *WOFStandardPlacesResult`
// thingy (and omit them from the JSON output) but today that is unnecessary
// (20170816/thisisaaronland)

func existentialFlag(i int64) flags.ExistentialFlag {
	fl, _ := existential.NewKnownUnknownFlag(i)
	return fl
}
