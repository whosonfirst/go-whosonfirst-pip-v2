package extras

// this is VERY VERY experimental still - basically we're going to decorate the
// final JSON output with extras data read from one or more SQLite databases -
// in time this will probably be updated to use go-whosonfirst-readwrite.Reader
// instances and some "S3 SELECT" -like for user-defined databases but not today
// (20171217/thisisaaronland)

// put another way - some sort of generic extras interface rather than explicitly
// requiring a SQLite database (20180303/thisisaaronland)

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-whosonfirst-spr"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	_ "log"
	"strings"
	"sync"
)

func AppendExtrasWithSPRResults(js []byte, results spr.StandardPlacesResults, paths []string, extras_db *database.SQLiteDatabase) ([]byte, error) {

	// okay - the thing to remember is that a this point we are working with
	// bytes (specifically the 'js' []byte variable) and we _not_ be working
	// with WOF-flavoured GeoJSON/SPR results so we can't know for certain which
	// property is being used to for the primary key - since we still have the
	// original 'results' variable, let's loop over it and create a simple index
	// based list of spr.Id() values that we can use as a pointer/lookup in
	// the AppendExtras* methods below (20180303/thisisaaronland)

	id_map := make([]string, len(results.Results()))

	for i, r := range results.Results() {
		id_map[i] = r.Id()
	}

	return AppendExtras(js, id_map, paths, extras_db)
}

func AppendExtras(js []byte, id_map []string, paths []string, extras_db *database.SQLiteDatabase) ([]byte, error) {

	conn, err := extras_db.Conn()

	if err != nil {
		return js, err
	}

	type update struct {
		Index int
		SPR   interface{}
	}

	done_ch := make(chan bool)
	update_ch := make(chan update)
	error_ch := make(chan error)

	rsp := gjson.GetBytes(js, "places")
	places := rsp.Array()

	count := len(places)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i, pl := range places {

		go func(ctx context.Context, idx int, pl gjson.Result) {

			defer func() {
				done_ch <- true
			}()

			select {

			case <-ctx.Done():
				// return
			default:
				id := id_map[idx] // see notes above
				raw := []byte(pl.Raw)

				updated, err := AppendExtrasToSPRBytes(raw, id, paths, conn)

				if err != nil {
					error_ch <- err
					return
				}

				if updated == nil {
					return
				}

				var spr interface{}
				err = json.Unmarshal(updated, &spr)

				if err != nil {
					error_ch <- err
					return
				}

				up := update{
					Index: idx,
					SPR:   spr,
				}

				update_ch <- up
			}
		}(ctx, i, pl)
	}

	mu := new(sync.Mutex)
	remaining := count

	for remaining > 0 {

		select {
		case <-done_ch:
			remaining -= 1
		case err := <-error_ch:
			return nil, err
		case up := <-update_ch:

			var err error

			mu.Lock()

			set_path := fmt.Sprintf("places.%d", up.Index)
			js, err = sjson.SetBytes(js, set_path, up.SPR)

			mu.Unlock()

			if err != nil {
				return nil, err
			}
		}
	}

	return js, nil
}

// there appears to be a bug in here (I think?) that prevents the CLI tools from exiting on a
// control-C event if we are using an extras DB that is a tempfile and that hasn't been indexed
// (for example if things were started using -mode spatialite) - I'm still not entirely sure
// about the cause... just the symptoms (20180308/thisisaaronland)

func AppendExtrasToSPRBytes(spr []byte, id string, extras []string, conn *sql.DB) ([]byte, error) {

	// apparently JSON_EXTRACT isn't available in go-sqlite yet?
	// 2017/12/17 20:07:00 420561633 no such function: JSON_EXTRACT
	// row := conn.QueryRow("SELECT JSON_EXTRACT(feature, '$.properties') FROM geojson WHERE id=?", id)

	// see also: https://github.com/whosonfirst/go-whosonfirst-pip-v2/issues/19

	row := conn.QueryRow("SELECT body FROM geojson WHERE id=?", id)

	var body []byte
	err := row.Scan(&body)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		// pass
	}

	for _, e := range extras {

		paths := make([]string, 0)

		if strings.HasSuffix(e, "*") || strings.HasSuffix(e, ":") {

			e = strings.Replace(e, "*", "", -1)

			props := gjson.GetBytes(body, "properties")

			for k, _ := range props.Map() {

				if strings.HasPrefix(k, e) {
					paths = append(paths, k)
				}
			}

		} else {
			paths = append(paths, e)
		}

		for _, p := range paths {

			// see above inre absence of JSON_EXTRACT function

			get_path := fmt.Sprintf("properties.%s", p)
			set_path := fmt.Sprintf("%s", p)

			v := gjson.GetBytes(body, get_path)

			/*
				log.Println("GET", id, get_path)
				log.Println("SET", id, set_path)
				log.Println("VALUE", v.Value())
			*/

			if v.Exists() {
				spr, err = sjson.SetBytes(spr, set_path, v.Value())
			} else {
				spr, err = sjson.SetBytes(spr, set_path, nil)
			}

			if err != nil {
				return nil, err
			}
		}
	}

	return spr, nil
}
