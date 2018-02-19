package index

// https://gist.github.com/simonw/91a1157d1f45ab305c6f48c4ca344de8

import (
	"github.com/skelterjohn/geom"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/geometry"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-pip"
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	"github.com/whosonfirst/go-whosonfirst-spr"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	// golog "log"
	"sync"
)

type SpatialiteIndex struct {
	Index
	Logger *log.WOFLogger
	cache  cache.Cache
}

func NewSpatialiteIndex(c cache.Cache) (Index, error) {

	db, err := database.NewDBWithDriver(*driver, *dsn)

	logger := log.SimpleWOFLogger("index")

	i := SpatialiteIndex{
		cache:  c,
		Logger: logger,
	}

	/*

		conn2 = sqlite3.connect('whole-earth-spatialite.db')
		conn2.enable_load_extension(True)
		conn2.execute("SELECT load_extension('/usr/local/lib/mod_spatialite.dylib');")
		conn2.execute("SELECT InitSpatialMetaData()")
		conn2.execute("CREATE TABLE whosonfirst (id INTEGER PRIMARY KEY, name TEXT, placetype INTEGER, properties TEXT)")
		conn2.execute("SELECT AddGeometryColumn('whosonfirst', 'geom', 2154, 'GEOMETRY', 'XY')")
		conn2.execute("SELECT CreateSpatialIndex('whosonfirst', 'geom')")

	*/

	return &i, nil
}

func (i *SpatialiteIndex) Cache() cache.Cache {

}

func (i *SpatialiteIndex) IndexFeature(f geojson.Feature) error {

	/*

	    for d in iterate_geojson():
	   sql = """
	   INSERT INTO whosonfirst
	       (id, name, placetype, properties, geom)
	   VALUES
	       (:id, :name, :placetype, :properties, GeomFromText(:geom, 2154))
	   """
	   params = {
	       'id': d['id'],
	       'name': d['name'],
	       'placetype': PlaceType[d['placetype']].value,
	       'properties': json.dumps(d['properties']),
	       'geom': shape(d['geometry']).wkt,
	   }
	   c.execute(sql, params

	*/
}

func GetIntersectsByCoord(geom.Coord, filter.Filter) (spr.StandardPlacesResults, error) {

	/*

	select
	  id, placetype, name, length(geom), properties
	from
	  whosonfirst
	where
	  within(GeomFromText('POINT(' || :longitude || ' ' || :latitude || ')'), geom)
	  and rowid in (
	        SELECT pkid FROM idx_whosonfirst_geom
	        where xmin < :longitude
	        and xmax > :longitude
	        and ymin < :latitude
	        and ymax > :latitude)
	order by placetype desc;

	*/

}

func GetCandidatesByCoord(geom.Coord) (*pip.GeoJSONFeatureCollection, error) {

}

func GetIntersectsByPath(geom.Path, filter.Filter) ([]spr.StandardPlacesResults, error) {

}
