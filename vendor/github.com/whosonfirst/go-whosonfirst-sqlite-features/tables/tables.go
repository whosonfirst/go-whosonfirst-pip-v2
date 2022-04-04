package tables

import (
	"context"
	"github.com/aaronland/go-sqlite"
)

type TableOptions struct {
	IndexAltFiles bool
}

type CommonTablesOptions struct {
	GeoJSON       *GeoJSONTableOptions // DEPRECATED
	IndexAltFiles bool
}

func CommonTablesWithDatabase(ctx context.Context, db sqlite.Database) ([]sqlite.Table, error) {

	geojson_opts, err := DefaultGeoJSONTableOptions()

	if err != nil {
		return nil, err
	}

	table_opts := &CommonTablesOptions{
		GeoJSON:       geojson_opts,
		IndexAltFiles: false,
	}

	return CommonTablesWithDatabaseAndOptions(ctx, db, table_opts)
}

func CommonTablesWithDatabaseAndOptions(ctx context.Context, db sqlite.Database, table_opts *CommonTablesOptions) ([]sqlite.Table, error) {

	to_index := make([]sqlite.Table, 0)

	var geojson_opts *GeoJSONTableOptions

	// table_opts.GeoJSON is deprecated but maintained for backwards compatbility
	// (20201224/thisisaaronland)

	if table_opts.GeoJSON != nil {
		geojson_opts = table_opts.GeoJSON
	} else {

		opts, err := DefaultGeoJSONTableOptions()

		if err != nil {
			return nil, err
		}

		opts.IndexAltFiles = table_opts.IndexAltFiles
		geojson_opts = opts
	}

	gt, err := NewGeoJSONTableWithDatabaseAndOptions(ctx, db, geojson_opts)

	if err != nil {
		return nil, err
	}

	to_index = append(to_index, gt)

	st_opts, err := DefaultSPRTableOptions()

	if err != nil {
		return nil, err
	}

	st_opts.IndexAltFiles = table_opts.IndexAltFiles

	st, err := NewSPRTableWithDatabaseAndOptions(ctx, db, st_opts)

	if err != nil {
		return nil, err
	}

	to_index = append(to_index, st)

	nm, err := NewNamesTableWithDatabase(ctx, db)

	if err != nil {
		return nil, err
	}

	to_index = append(to_index, nm)

	an, err := NewAncestorsTableWithDatabase(ctx, db)

	if err != nil {
		return nil, err
	}

	to_index = append(to_index, an)

	cn, err := NewConcordancesTableWithDatabase(ctx, db)

	if err != nil {
		return nil, err
	}

	to_index = append(to_index, cn)

	return to_index, nil
}

func SpatialTablesWithDatabase(ctx context.Context, db sqlite.Database) ([]sqlite.Table, error) {

	to_index := make([]sqlite.Table, 0)

	st, err := NewGeometriesTableWithDatabase(ctx, db)

	if err != nil {
		return nil, err
	}

	to_index = append(to_index, st)
	return to_index, nil
}

func PointInPolygonTablesWithDatabase(ctx context.Context, db sqlite.Database) ([]sqlite.Table, error) {

	to_index, err := SpatialTablesWithDatabase(ctx, db)

	if err != nil {
		return nil, err
	}

	gt, err := NewGeoJSONTableWithDatabase(ctx, db)

	if err != nil {
		return nil, err
	}

	to_index = append(to_index, gt)

	return to_index, nil
}

func SearchTablesWithDatabase(ctx context.Context, db sqlite.Database) ([]sqlite.Table, error) {

	opts := &TableOptions{
		IndexAltFiles: false,
	}

	return SearchTablesWithDatabaseAndOptions(ctx, db, opts)
}

func SearchTablesWithDatabaseAndOptions(ctx context.Context, db sqlite.Database, opts *TableOptions) ([]sqlite.Table, error) {

	to_index := make([]sqlite.Table, 0)

	st, err := NewSearchTableWithDatabase(ctx, db)

	if err != nil {
		return nil, err
	}

	to_index = append(to_index, st)
	return to_index, nil
}

func RTreeTablesWithDatabase(ctx context.Context, db sqlite.Database) ([]sqlite.Table, error) {

	opts := &TableOptions{
		IndexAltFiles: false,
	}

	return RTreeTablesWithDatabaseAndOptions(ctx, db, opts)
}

func RTreeTablesWithDatabaseAndOptions(ctx context.Context, db sqlite.Database, opts *TableOptions) ([]sqlite.Table, error) {

	// https://github.com/whosonfirst/go-whosonfirst-spatial-sqlite#databases

	to_index := make([]sqlite.Table, 0)

	rtree_opts, err := DefaultRTreeTableOptions()

	if err != nil {
		return nil, err
	}

	rtree_opts.IndexAltFiles = opts.IndexAltFiles

	rt, err := NewRTreeTableWithDatabaseAndOptions(ctx, db, rtree_opts)

	if err != nil {
		return nil, err
	}

	to_index = append(to_index, rt)

	sprt_opts, err := DefaultSPRTableOptions()

	if err != nil {
		return nil, err
	}

	sprt_opts.IndexAltFiles = opts.IndexAltFiles

	sprt, err := NewSPRTableWithDatabaseAndOptions(ctx, db, sprt_opts)

	if err != nil {
		return nil, err
	}

	to_index = append(to_index, sprt)

	props_opts, err := DefaultPropertiesTableOptions()

	if err != nil {
		return nil, err
	}

	props_opts.IndexAltFiles = opts.IndexAltFiles

	props, err := NewPropertiesTableWithDatabaseAndOptions(ctx, db, props_opts)

	if err != nil {
		return nil, err
	}

	to_index = append(to_index, props)

	geom_opts, err := DefaultGeometryTableOptions()

	if err != nil {
		return nil, err
	}

	geom_opts.IndexAltFiles = opts.IndexAltFiles

	geom, err := NewGeometryTableWithDatabaseAndOptions(ctx, db, geom_opts)

	if err != nil {
		return nil, err
	}

	to_index = append(to_index, geom)

	return to_index, nil
}
