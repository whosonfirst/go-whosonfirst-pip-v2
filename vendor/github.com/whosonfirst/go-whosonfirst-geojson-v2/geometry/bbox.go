package geometry

import (
	"github.com/skelterjohn/geom"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
)

type Bboxes struct {
	geojson.BoundingBoxes `json:",omitempty"`
	BBoxesBounds          []*geom.Rect `json:"bounds"`
	BBoxesMBR             geom.Rect    `json:"mbr"`
}

func (b Bboxes) Bounds() []*geom.Rect {
	return b.BBoxesBounds
}

func (b Bboxes) MBR() geom.Rect {
	return b.BBoxesMBR
}

func BoundingBoxesForFeature(f geojson.Feature) (geojson.BoundingBoxes, error) {

	polys, err := PolygonsForFeature(f)

	if err != nil {
		return nil, err
	}

	mbr := geom.NilRect()
	bounds := make([]*geom.Rect, 0)

	for _, poly := range polys {

		ext := poly.ExteriorRing()
		b := ext.Path.Bounds()

		mbr.ExpandToContainRect(*b)
		bounds = append(bounds, b)
	}

	wb := Bboxes{
		BBoxesBounds: bounds,
		BBoxesMBR:    mbr,
	}

	return wb, nil
}
