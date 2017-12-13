package util

import (
	"encoding/json"
	"github.com/whosonfirst/go-whosonfirst-spr"
	"strconv"
	"strings"
)

func SPRToMap(s spr.StandardPlacesResult) (map[string]string, error) {

	attrs := make(map[string]string)

	enc, err := json.Marshal(s)

	if err != nil {
		return attrs, err
	}

	var tmp map[string]interface{}

	err = json.Unmarshal(enc, &tmp)

	if err != nil {
		return attrs, err
	}

	for k, v := range tmp {

		var str_v string

		switch t := v.(type) {

		case int64:
			str_v = strconv.FormatInt(t, 10)
		case float64:
			str_v = strconv.FormatFloat(t, 'f', -1, 64)
		case string:
			str_v = t
		default:

			tmp := v.([]interface{})
			ids := make([]string, 0)

			for _, i := range tmp {

				var i64 int64

				switch it := i.(type) {

				case float64:
					i64 = int64(it)
				case int64:
					i64 = it
				default:
					i64 = 0
				}

				ids = append(ids, strconv.FormatInt(i64, 10))
			}

			str_v = strings.Join(ids, ",")
		}

		attrs[k] = str_v
	}

	return attrs, nil
}
