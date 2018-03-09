package tables

import (
	"github.com/whosonfirst/go-whosonfirst-sqlite"
)

func CommonTablesWithDatabase(db sqlite.Database) ([]sqlite.Table, error) {

	to_index := make([]sqlite.Table, 0)

	gt, err := NewGeoJSONTableWithDatabase(db)

	if err != nil {
		return nil, err
	}

	to_index = append(to_index, gt)

	st, err := NewSPRTableWithDatabase(db)

	if err != nil {
		return nil, err
	}

	to_index = append(to_index, st)

	nm, err := NewNamesTableWithDatabase(db)

	if err != nil {
		return nil, err
	}

	to_index = append(to_index, nm)

	an, err := NewAncestorsTableWithDatabase(db)

	if err != nil {
		return nil, err
	}

	to_index = append(to_index, an)

	cn, err := NewConcordancesTableWithDatabase(db)

	if err != nil {
		return nil, err
	}

	to_index = append(to_index, cn)

	return to_index, nil
}

func SpatialTablesWithDatabase(db sqlite.Database) ([]sqlite.Table, error) {

	to_index := make([]sqlite.Table, 0)

	st, err := NewGeometriesTableWithDatabase(db)

	if err != nil {
		return nil, err
	}

	to_index = append(to_index, st)
	return to_index, nil
}

func PointInPolygonTablesWithDatabase(db sqlite.Database) ([]sqlite.Table, error) {

	to_index, err := SpatialTablesWithDatabase(db)

	if err != nil {
		return nil, err
	}

	gt, err := NewGeoJSONTableWithDatabase(db)

	if err != nil {
		return nil, err
	}

	to_index = append(to_index, gt)

	return to_index, nil
}

func SearchTablesWithDatabase(db sqlite.Database) ([]sqlite.Table, error) {

	to_index := make([]sqlite.Table, 0)

	st, err := NewSearchTableWithDatabase(db)

	if err != nil {
		return nil, err
	}

	to_index = append(to_index, st)
	return to_index, nil
}
