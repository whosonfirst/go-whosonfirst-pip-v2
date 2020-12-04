package index

import (
	"context"
	"errors"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-log"
	"io"
	"net/url"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const (
	STDIN = "STDIN"
)

var (
	driversMu sync.RWMutex
	drivers   = make(map[string]Driver)
)

type IndexerFunc func(ctx context.Context, fh io.Reader, args ...interface{}) error

type IndexerContextKey string

type Indexer struct {
	Driver  Driver
	Func    IndexerFunc
	Logger  *log.WOFLogger
	Indexed int64
	count   int64
}

// used by the IndexGit stuff
// https://godoc.org/gopkg.in/src-d/go-git.v4/plumbing/protocol/packp/sideband#Progress

/*
type WOFLoggerProgress struct {
	sideband.Progress
	logger *log.WOFLogger
}

func (p *WOFLoggerProgress) Write(msg []byte) (int, error) {
	p.logger.Status(string(msg))
	return -1, nil
}
*/

func Register(name string, driver Driver) {

	driversMu.Lock()
	defer driversMu.Unlock()

	if driver == nil {
		panic("sql: Register driver is nil")

	}

	if _, dup := drivers[name]; dup {
		panic("index: Register called twice for driver " + name)
	}

	drivers[name] = driver
}

func unregisterAllDrivers() {
	driversMu.Lock()
	defer driversMu.Unlock()
	drivers = make(map[string]Driver)
}

func Drivers() []string {

	driversMu.RLock()
	defer driversMu.RUnlock()

	var list []string

	for name := range drivers {
		list = append(list, name)
	}

	sort.Strings(list)
	return list
}

func Modes() []string {

	return Drivers()
}

func ContextForPath(path string) (context.Context, error) {

	ctx := AssignPathContext(context.Background(), path)
	return ctx, nil
}

func AssignPathContext(ctx context.Context, path string) context.Context {

	key := IndexerContextKey("path")
	return context.WithValue(ctx, key, path)
}

func PathForContext(ctx context.Context) (string, error) {

	k := IndexerContextKey("path")
	path := ctx.Value(k)

	if path == nil {
		return "", errors.New("path is not set")
	}

	return path.(string), nil
}

func NewIndexer(dsn string, f IndexerFunc) (*Indexer, error) {

	driversMu.Lock()
	defer driversMu.Unlock()

	u, err := url.Parse(dsn)

	if err != nil {
		return nil, err
	}

	name := u.Scheme

	// this is here for backwards compatibility

	if name == "" {

		dsn = fmt.Sprintf("%s://", dsn)

		u, err := url.Parse(dsn)

		if err != nil {
			return nil, err
		}

		name = u.Scheme
	}

	driver, ok := drivers[name]

	if !ok {
		return nil, errors.New("Unknown driver")
	}

	err = driver.Open(dsn)

	if err != nil {
		return nil, err
	}

	logger := log.SimpleWOFLogger("index")

	i := Indexer{
		Driver:  driver,
		Func:    f,
		Logger:  logger,
		Indexed: 0,
		count:   0,
	}

	return &i, nil
}

func (i *Indexer) Index(ctx context.Context, paths ...string) error {

	t1 := time.Now()

	defer func() {
		t2 := time.Since(t1)
		i.Logger.Status("time to index paths (%d) %v", len(paths), t2)
	}()

	i.increment()
	defer i.decrement()

	counter_func := func(ctx context.Context, fh io.Reader, args ...interface{}) error {
		defer atomic.AddInt64(&i.Indexed, 1)
		return i.Func(ctx, fh, args...)
	}

	for _, path := range paths {

		select {
		case <-ctx.Done():
			break
		default:
			// pass
		}

		err := i.Driver.IndexURI(ctx, counter_func, path)

		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Indexer) IndexPaths(paths []string, args ...interface{}) error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return i.Index(ctx, paths...)
}

func (i *Indexer) IndexPath(path string, args ...interface{}) error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return i.Index(ctx, path)
}

func (i *Indexer) IsIndexing() bool {

	if atomic.LoadInt64(&i.count) > 0 {
		return true
	}

	return false
}

func (i *Indexer) increment() {
	atomic.AddInt64(&i.count, 1)
}

func (i *Indexer) decrement() {
	atomic.AddInt64(&i.count, -1)
}
