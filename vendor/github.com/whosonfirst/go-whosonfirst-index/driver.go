package index

import (
	"context"
)

type Driver interface {
	Open(string) error
	IndexURI(context.Context, IndexerFunc, string) error
}
