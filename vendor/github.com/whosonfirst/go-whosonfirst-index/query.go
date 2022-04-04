package index

import (
	"bytes"
	"context"
	"github.com/aaronland/go-json-query"
	"io"
	"io/ioutil"
)

func NewCallbackWithQuerySet(ctx context.Context, indexer_cb IndexerFunc, qs *query.QuerySet) (IndexerFunc, error) {

	if qs == nil {
		return indexer_cb, nil
	}

	if len(qs.Queries) == 0 {
		return indexer_cb, nil
	}

	query_cb := func(ctx context.Context, fh io.Reader, args ...interface{}) error {

		body, err := ioutil.ReadAll(fh)

		if err != nil {
			return err
		}

		matches, err := query.Matches(ctx, qs, body)

		if err != nil {
			return err
		}

		if !matches {
			return nil
		}

		br := bytes.NewReader(body)
		return indexer_cb(ctx, br, args...)
	}

	return query_cb, nil
}

func NewIndexerWithQuerySet(ctx context.Context, indexer_uri string, indexer_cb IndexerFunc, qs *query.QuerySet) (*Indexer, error) {

	query_cb, err := NewCallbackWithQuerySet(ctx, indexer_cb, qs)

	if err != nil {
		return nil, err
	}

	return NewIndexer(indexer_uri, query_cb)
}
