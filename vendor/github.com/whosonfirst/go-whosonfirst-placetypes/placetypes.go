package placetypes

import (
	"encoding/json"
	"errors"
	"github.com/whosonfirst/go-whosonfirst-placetypes/placetypes"
	"log"
	"strconv"
)

type WOFPlacetypeName struct {
	Lang string `json:"language"`
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type WOFPlacetypeAltNames map[string][]string

type WOFPlacetype struct {
	Id     int64   `json:"id"`
	Name   string  `json:"name"`
	Role   string  `json:"role"`
	Parent []int64 `json:"parent"`
	// AltNames []WOFPlacetypeAltNames		`json:"names"`
}

type WOFPlacetypeSpecification map[string]WOFPlacetype

var specification *WOFPlacetypeSpecification

func init() {

	var err error

	specification, err = Spec()

	if err != nil {
		log.Fatal("Failed to parse specification", err)
	}
}

func Spec() (*WOFPlacetypeSpecification, error) {

	var spec WOFPlacetypeSpecification
	err := json.Unmarshal([]byte(placetypes.Specification), &spec)

	if err != nil {
		return nil, err
	}

	return &spec, nil
}

func IsValidPlacetype(name string) bool {

	for _, pt := range *specification {

		if pt.Name == name {
			return true
		}
	}

	return false
}

func IsValidPlacetypeId(id int64) bool {

	for str_id, _ := range *specification {

		pt_id, err := strconv.Atoi(str_id)

		if err != nil {
			continue
		}

		pt_id64 := int64(pt_id)

		if pt_id64 == id {
			return true
		}
	}

	return false
}

func GetPlacetypeByName(name string) (*WOFPlacetype, error) {

	for str_id, pt := range *specification {

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

func GetPlacetypeById(id int64) (*WOFPlacetype, error) {

	for str_id, pt := range *specification {

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

func Children(pt *WOFPlacetype) []*WOFPlacetype {

	children := make([]*WOFPlacetype, 0)

	for _, details := range *specification {

		for _, pid := range details.Parent {

			if pid == pt.Id {
				child_pt, _ := GetPlacetypeByName(details.Name)
				children = append(children, child_pt)
			}
		}
	}

	return sortChildren(pt, children)
}

func sortChildren(pt *WOFPlacetype, all []*WOFPlacetype) []*WOFPlacetype {

	kids := make([]*WOFPlacetype, 0)
	grandkids := make([]*WOFPlacetype, 0)

	for _, other := range all {

		is_grandkid := false

		for _, pid := range other.Parent {

			for _, p := range all {

				if pid == p.Id {
					is_grandkid = true
					break
				}
			}

			if is_grandkid {
				break
			}
		}

		if is_grandkid {
			grandkids = append(grandkids, other)
		} else {
			kids = append(kids, other)
		}
	}

	if len(grandkids) > 0 {
		grandkids = sortChildren(pt, grandkids)
	}

	for _, k := range grandkids {
		kids = append(kids, k)
	}

	return kids
}

func Descendants(pt *WOFPlacetype) []*WOFPlacetype {
	return DescendantsForRoles(pt, []string{"common"})
}

func DescendantsForRoles(pt *WOFPlacetype, roles []string) []*WOFPlacetype {

	descendants := make([]*WOFPlacetype, 0)
	descendants = fetchDescendants(pt, roles, descendants)

	return descendants
}

func fetchDescendants(pt *WOFPlacetype, roles []string, descendants []*WOFPlacetype) []*WOFPlacetype {

	grandkids := make([]*WOFPlacetype, 0)

	for _, kid := range Children(pt) {

		descendants = appendPlacetype(kid, roles, descendants)

		for _, grandkid := range Children(kid) {
			grandkids = appendPlacetype(grandkid, roles, grandkids)
		}
	}

	for _, k := range grandkids {
		descendants = appendPlacetype(k, roles, descendants)
		descendants = fetchDescendants(k, roles, descendants)
	}

	return descendants
}

func appendPlacetype(pt *WOFPlacetype, roles []string, others []*WOFPlacetype) []*WOFPlacetype {

	do_append := true

	for _, o := range others {

		if pt.Id == o.Id {
			do_append = false
			break
		}
	}

	if !do_append {
		return others
	}

	has_role := false

	for _, r := range roles {

		if pt.Role == r {
			has_role = true
			break
		}
	}

	if !has_role {
		return others
	}

	others = append(others, pt)
	return others
}

func Ancestors(pt *WOFPlacetype) []*WOFPlacetype {
	return AncestorsForRoles(pt, []string{"common"})
}

func AncestorsForRoles(pt *WOFPlacetype, roles []string) []*WOFPlacetype {

	ancestors := make([]*WOFPlacetype, 0)
	ancestors = fetchAncestors(pt, roles, ancestors)

	return ancestors
}

func fetchAncestors(pt *WOFPlacetype, roles []string, ancestors []*WOFPlacetype) []*WOFPlacetype {

	for _, id := range pt.Parent {

		parent, _ := GetPlacetypeById(id)

		role_ok := false

		for _, r := range roles {

			if r == parent.Role {
				role_ok = true
				break
			}
		}

		if !role_ok {
			continue
		}

		append_ok := true

		for _, a := range ancestors {

			if a.Id == parent.Id {
				append_ok = false
				break
			}
		}

		if append_ok {

			has_grandparent := false
			offset := -1

			for _, gpid := range parent.Parent {

				for idx, a := range ancestors {

					if a.Id == gpid {
						offset = idx
						has_grandparent = true
						break
					}
				}

				if has_grandparent {
					break
				}
			}

			// log.Printf("APPEND %s < %s GP: %t (%d)\n", parent.Name, pt.Name, has_grandparent, offset)

			if has_grandparent {

				// log.Println("WTF 1", len(ancestors))

				tail := ancestors[offset+1:]
				ancestors = ancestors[0:offset]

				ancestors = append(ancestors, parent)

				for _, a := range tail {
					ancestors = append(ancestors, a)
				}

			} else {
				ancestors = append(ancestors, parent)
			}
		}

		ancestors = fetchAncestors(parent, roles, ancestors)
	}

	return ancestors
}
