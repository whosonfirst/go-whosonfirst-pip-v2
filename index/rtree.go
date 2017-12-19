package index

import (
	"github.com/dhconnelly/rtreego"
	"github.com/skelterjohn/geom"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/geometry"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-pip"
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	"github.com/whosonfirst/go-whosonfirst-spr"
	// golog "log"
	"sync"
)

type RTreeIndex struct {
	Index
	Logger *log.WOFLogger
	rtree  *rtreego.Rtree
	cache  cache.Cache
	mu     *sync.RWMutex
}

type RTreeSpatialIndex struct {
	bounds *rtreego.Rect
	Id     string
}

func (sp RTreeSpatialIndex) Bounds() *rtreego.Rect {
	return sp.bounds
}

type RTreeResults struct {
	spr.StandardPlacesResults `json:",omitempty"`
	Places                    []spr.StandardPlacesResult `json:"places"`
}

func (r *RTreeResults) Results() []spr.StandardPlacesResult {
	return r.Places
}

func NewRTreeIndex(c cache.Cache) (*RTreeIndex, error) {

	logger := log.SimpleWOFLogger("index")

	rtree := rtreego.NewTree(2, 25, 50)

	mu := new(sync.RWMutex)

	index := RTreeIndex{
		Logger: logger,
		rtree:  rtree,
		cache:  c,
		mu:     mu,
	}

	return &index, nil
}

func (r *RTreeIndex) Cache() cache.Cache {
	return r.cache
}

func (r *RTreeIndex) IndexFeature(f geojson.Feature) error {

	str_id := f.Id()

	bboxes, err := f.BoundingBoxes()

	if err != nil {
		return err
	}

	fc, err := cache.NewFeatureCache(f)

	if err != nil {
		return err
	}

	err = r.cache.Set(str_id, fc)

	if err != nil {
		return err
	}

	for _, bbox := range bboxes.Bounds() {

		sw := bbox.Min
		ne := bbox.Max

		llat := ne.Y - sw.Y
		llon := ne.X - sw.X

		pt := rtreego.Point{sw.X, sw.Y}
		rect, err := rtreego.NewRect(pt, []float64{llon, llat})

		if err != nil {
			return err
		}

		r.Logger.Status("index %s %v", str_id, rect)

		sp := RTreeSpatialIndex{
			bounds: rect,
			Id:     str_id,
		}

		r.mu.Lock()
		r.rtree.Insert(&sp)
		r.mu.Unlock()
	}

	return nil
}

func (r *RTreeIndex) GetIntersectsByPath(path geom.Path, filters filter.Filter) ([]spr.StandardPlacesResults, error) {

	type Candidates struct {
		Index int
		SPR   spr.StandardPlacesResults
	}

	candidates_ch := make(chan Candidates)
	error_ch := make(chan error)
	done_ch := make(chan bool)

	pending := path.Length()
	results := make([]spr.StandardPlacesResults, pending)

	for i, c := range path.Vertices() {

		// see the way we're passing 'i' around as an index - that's so we rebuild
		// the result sets in the same order that the coords were passed in - that
		// part is important (20170927/thisisaaronland)

		go func(idx int, c geom.Coord, f filter.Filter, candidates_ch chan Candidates, error_ch chan error, done_ch chan bool) {

			defer func() {
				done_ch <- true
			}()

			intersects, err := r.GetIntersectsByCoord(c, f)

			if err != nil {
				error_ch <- err
				return
			}

			candidates := Candidates{
				Index: idx,
				SPR:   intersects,
			}

			candidates_ch <- candidates

		}(i, c, filters, candidates_ch, error_ch, done_ch)

	}

	// notes (20170927/thisisaaronland)
	// 1. kill remaining goroutines if err

	for pending > 0 {

		select {

		case err := <-error_ch:
			return nil, err
		case candidates := <-candidates_ch:
			results[candidates.Index] = candidates.SPR
		case <-done_ch:
			pending -= 1
		}
	}

	return results, nil
}

func (r *RTreeIndex) GetIntersectsByCoord(coord geom.Coord, filters filter.Filter) (spr.StandardPlacesResults, error) {

	// to do: timings that don't slow everything down the way
	// go-whosonfirst-timer does now (20170915/thisisaaronland)

	rows, err := r.getIntersectsByCoord(coord)

	if err != nil {
		return nil, err
	}

	rsp, err := r.inflateResults(coord, filters, rows)

	if err != nil {
		return nil, err
	}

	return rsp, err
}

func (r *RTreeIndex) GetCandidatesByCoord(coord geom.Coord) (*pip.GeoJSONFeatureCollection, error) {

	intersects, err := r.getIntersectsByCoord(coord)

	if err != nil {
		return nil, err
	}

	features := make([]pip.GeoJSONFeature, 0)

	for _, raw := range intersects {

		spatial := raw.(*RTreeSpatialIndex)
		str_id := spatial.Id

		props := map[string]interface{}{
			"id": str_id,
		}

		b := spatial.Bounds()

		swlon := b.PointCoord(0)
		swlat := b.PointCoord(1)

		nelon := swlon + b.LengthsCoord(0)
		nelat := swlat + b.LengthsCoord(1)

		sw := pip.GeoJSONPoint{swlon, swlat}
		nw := pip.GeoJSONPoint{swlon, nelat}
		ne := pip.GeoJSONPoint{nelon, nelat}
		se := pip.GeoJSONPoint{nelon, swlat}

		ring := pip.GeoJSONRing{sw, nw, ne, se, sw}
		poly := pip.GeoJSONPolygon{ring}
		multi := pip.GeoJSONMultiPolygon{poly}

		geom := pip.GeoJSONGeometry{
			Type:        "MultiPolygon",
			Coordinates: multi,
		}

		feature := pip.GeoJSONFeature{
			Type:       "Feature",
			Properties: props,
			Geometry:   geom,
		}

		features = append(features, feature)
	}

	fc := pip.GeoJSONFeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}

	return &fc, nil
}

func (r *RTreeIndex) getIntersectsByCoord(coord geom.Coord) ([]rtreego.Spatial, error) {

	lat := coord.Y
	lon := coord.X

	pt := rtreego.Point{lon, lat}
	rect, err := rtreego.NewRect(pt, []float64{0.0001, 0.0001}) // how small can I make this?

	if err != nil {
		return nil, err
	}

	return r.getIntersectsByRect(rect)
}

func (r *RTreeIndex) getIntersectsByRect(rect *rtreego.Rect) ([]rtreego.Spatial, error) {

	// to do: timings that don't slow everything down the way
	// go-whosonfirst-timer does now (20170915/thisisaaronland)

	results := r.rtree.SearchIntersect(rect)
	return results, nil
}

func (r *RTreeIndex) inflateResults(c geom.Coord, f filter.Filter, possible []rtreego.Spatial) (spr.StandardPlacesResults, error) {

	// to do: timings that don't slow everything down the way
	// go-whosonfirst-timer does now (20170915/thisisaaronland)

	rows := make([]spr.StandardPlacesResult, 0)
	seen := make(map[string]bool)

	mu := new(sync.RWMutex)
	wg := new(sync.WaitGroup)

	for _, row := range possible {

		sp := row.(*RTreeSpatialIndex)
		wg.Add(1)

		go func(sp *RTreeSpatialIndex) {

			defer wg.Done()

			str_id := sp.Id

			mu.RLock()
			_, ok := seen[str_id]
			mu.RUnlock()

			if ok {
				return
			}

			mu.Lock()
			seen[str_id] = true
			mu.Unlock()

			fc, err := r.cache.Get(str_id)

			if err != nil {
				r.Logger.Error("failed to retrieve cache for %s, because %s", str_id, err)
				return
			}

			s := fc.SPR()

			err = filter.FilterSPR(f, s)

			if err != nil {
				r.Logger.Debug("SKIP %s because filter error %s", str_id, err)
				return
			}

			p := fc.Polygons()

			contains, err := geometry.PolygonsContainsCoord(p, c)

			if err != nil {
				r.Logger.Error("failed to calculate intersection for %s, because %s", str_id, err)
				return
			}

			if !contains {
				r.Logger.Debug("SKIP %s because does not contain coord (%v)", str_id, c)
				return
			}

			// r.Logger.Status("APPEND %s to result set", str_id)

			mu.Lock()
			rows = append(rows, s)
			mu.Unlock()

		}(sp)
	}

	wg.Wait()

	rs := RTreeResults{
		Places: rows,
	}

	return &rs, nil
}
