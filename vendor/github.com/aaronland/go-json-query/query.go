package query

import (
	"context"
	"github.com/tidwall/gjson"
	_ "log"
	"regexp"
)

const QUERYSET_MODE_ANY string = "ANY"
const QUERYSET_MODE_ALL string = "ALL"

type QuerySet struct {
	Queries []*Query
	Mode    string
}

type Query struct {
	Path  string
	Match *regexp.Regexp
}

func Matches(ctx context.Context, qs *QuerySet, body []byte) (bool, error) {

	select {
	case <-ctx.Done():
		return false, nil
	default:
		// pass
	}

	queries := qs.Queries
	mode := qs.Mode

	tests := len(queries)
	matches := 0

	for _, q := range queries {

		rsp := gjson.GetBytes(body, q.Path)

		if !rsp.Exists() {

			if mode == QUERYSET_MODE_ALL {
				break
			}
		}

		for _, r := range rsp.Array() {

			if q.Match.MatchString(r.String()) {

				matches += 1

				if mode == QUERYSET_MODE_ANY {
					break
				}
			}
		}

		if mode == QUERYSET_MODE_ANY && matches > 0 {
			break
		}
		
	}

	if mode == QUERYSET_MODE_ALL {

		if matches < tests {
			return false, nil
		}
	}

	if matches == 0 {
		return false, nil
	}

	return true, nil
}
