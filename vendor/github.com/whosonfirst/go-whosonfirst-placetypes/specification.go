package placetypes

import (
	"encoding/json"
	"errors"
	"github.com/whosonfirst/go-whosonfirst-placetypes/placetypes"
	"io"
	"io/ioutil"
	"strconv"
	"sync"
)

type WOFPlacetypeSpecification struct {
	catalog map[string]WOFPlacetype
	mu      *sync.RWMutex
}

func DefaultWOFPlacetypeSpecification() (*WOFPlacetypeSpecification, error) {
	return NewWOFPlacetypeSpecification([]byte(placetypes.Specification))
}

func NewWOFPlacetypeSpecificationWithReader(r io.Reader) (*WOFPlacetypeSpecification, error) {

	body, err := ioutil.ReadAll(r)

	if err != nil {
		return nil, err
	}

	return NewWOFPlacetypeSpecification(body)
}

func NewWOFPlacetypeSpecification(body []byte) (*WOFPlacetypeSpecification, error) {

	var catalog map[string]WOFPlacetype
	err := json.Unmarshal(body, &catalog)

	if err != nil {
		return nil, err
	}

	mu := new(sync.RWMutex)

	spec := &WOFPlacetypeSpecification{
		catalog: catalog,
		mu:      mu,
	}

	return spec, nil
}

func (spec *WOFPlacetypeSpecification) GetPlacetypeByName(name string) (*WOFPlacetype, error) {

	// spec.mu.RLock()
	// defer spec.mu.RUnlock()

	for str_id, pt := range spec.catalog {

		if pt.Name == name {

			pt_id, err := strconv.Atoi(str_id)

			if err != nil {
				continue
			}

			pt_id64 := int64(pt_id)

			pt.Id = pt_id64
			return &pt, nil
		}
	}

	return nil, errors.New("Invalid placetype")
}

func (spec *WOFPlacetypeSpecification) GetPlacetypeById(id int64) (*WOFPlacetype, error) {

	// spec.mu.RLock()
	// defer spec.mu.RUnlock()

	for str_id, pt := range spec.catalog {

		pt_id, err := strconv.Atoi(str_id)

		if err != nil {
			continue
		}

		pt_id64 := int64(pt_id)

		if pt_id64 == id {
			pt.Id = pt_id64
			return &pt, nil
		}
	}

	return nil, errors.New("Invalid placetype")
}

func (spec *WOFPlacetypeSpecification) AppendPlacetypeSpecification(other_spec *WOFPlacetypeSpecification) error {

	for _, pt := range other_spec.Catalog() {

		err := spec.AppendPlacetype(pt)

		if err != nil {
			return err
		}
	}

	return nil
}

func (spec *WOFPlacetypeSpecification) AppendPlacetype(pt WOFPlacetype) error {

	spec.mu.Lock()
	defer spec.mu.Unlock()

	existing_pt, _ := spec.GetPlacetypeById(pt.Id)

	if existing_pt != nil {
		return errors.New("Placetype ID already registered")
	}

	existing_pt, _ = spec.GetPlacetypeByName(pt.Name)

	if existing_pt != nil {
		return errors.New("Placetype name already registered")
	}

	for _, pid := range pt.Parent {

		_, err := spec.GetPlacetypeById(pid)

		if err != nil {
			return err
		}
	}

	str_id := strconv.FormatInt(pt.Id, 10)
	spec.catalog[str_id] = pt
	return nil
}

func (spec *WOFPlacetypeSpecification) Catalog() map[string]WOFPlacetype {
	return spec.catalog
}
