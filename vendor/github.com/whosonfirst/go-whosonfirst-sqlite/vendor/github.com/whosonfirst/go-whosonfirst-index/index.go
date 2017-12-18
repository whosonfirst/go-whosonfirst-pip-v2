package index

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/whosonfirst/go-whosonfirst-crawl"
	"github.com/whosonfirst/go-whosonfirst-csv"
	"github.com/whosonfirst/go-whosonfirst-log"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"		
)

const (
	STDIN = "STDIN"
)

type IndexerFunc func(fh io.Reader, ctx context.Context, args ...interface{}) error

type IndexerContextKey string

type Indexer struct {
	Mode    string
	Func    IndexerFunc
	Logger  *log.WOFLogger
	Indexed int64
	count   int64
}

func Modes() []string {

	return []string{
		"directory",
		"feature",
		"feature-collection",
		"files",
		"geojson-ls",
		"meta",
		"path",
		"repo",
	}
}

func ContextForPath(path string) (context.Context, error) {

	key := IndexerContextKey("path")
	ctx := context.WithValue(context.Background(), key, path)

	return ctx, nil
}

func PathForContext(ctx context.Context) (string, error) {

	k := IndexerContextKey("path")
	path := ctx.Value(k)

	if path == nil {
		return "", errors.New("path is not set")
	}

	return path.(string), nil
}

func NewIndexer(mode string, f IndexerFunc) (*Indexer, error) {

	logger := log.SimpleWOFLogger("index")

	i := Indexer{
		Mode:    mode,
		Func:    f,
		Logger:  logger,
		Indexed: 0,
		count:   0,
	}

	return &i, nil
}

func (i *Indexer) IndexPaths(paths []string, args ...interface{}) error {

	t1 := time.Now()

	defer func(){
		t2 := time.Since(t1)
		i.Logger.Status("time to index paths (%d) %v", len(paths), t2)		   
	}()
	
	i.increment()
	defer i.decrement()

	for _, path := range paths {

		err := i.IndexPath(path, args...)

		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Indexer) IndexPath(path string, args ...interface{}) error {

	t1 := time.Now()

	defer func(){
		t2 := time.Since(t1)
		i.Logger.Status("time to index path '%s' %v", path, t2)
	}()

	i.increment()
	defer i.decrement()

	i.Logger.Debug("index %s in %s mode", path, i.Mode)

	if i.Mode == "directory" {

		return i.IndexDirectory(path, args...)

	} else if i.Mode == "feature" {

		return i.IndexFile(path, args...)

	} else if i.Mode == "feature-collection" {

		return i.IndexGeoJSONFeatureCollection(path, args...)

	} else if i.Mode == "filelist" {

		return i.IndexFileList(path, args...)

	} else if i.Mode == "files" {

		return i.IndexFile(path, args...)

	} else if i.Mode == "geojson-ls" {

		return i.IndexGeoJSONLS(path, args...)

	} else if i.Mode == "meta" {

		// please refactor all of this in to something... better
		// (20170823/thisisaaronland)

		parts := strings.Split(path, ":")

		if len(parts) == 1 {

			abs_root, err := filepath.Abs(parts[0])

			if err != nil {
				return err
			}

			meta_root := filepath.Dir(abs_root)
			repo_root := filepath.Dir(meta_root)
			data_root := filepath.Join(repo_root, "data")

			parts = append(parts, data_root)
		}

		if len(parts) != 2 {
			return errors.New("Invalid path declaration for a meta file")
		}

		for _, p := range parts {

			if p == STDIN {
				continue
			}

			_, err := os.Stat(p)

			if os.IsNotExist(err) {
				return errors.New("Path does not exist")
			}
		}

		meta_file := parts[0]
		data_root := parts[1]

		return i.IndexMetaFile(meta_file, data_root, args...)

	} else if i.Mode == "repo" {

		abs_path, err := filepath.Abs(path)

		if err != nil {
			return err
		}

		data := filepath.Join(abs_path, "data")

		_, err = os.Stat(data)

		if err != nil {
			return err
		}

		return i.IndexDirectory(data, args...)

	} else {

		return errors.New("Invalid indexer")
	}

}

func (i *Indexer) IndexFile(path string, args ...interface{}) error {

	t1 := time.Now()

	defer func(){
		t2 := time.Since(t1)
		i.Logger.Status("time to index file '%s' %v", path, t2)
	}()

	i.increment()
	defer i.decrement()

	fh, err := i.readerFromPath(path)

	if err != nil {
		return err
	}

	defer fh.Close()

	return i.process(fh, path, args...)
}

func (i *Indexer) IndexDirectory(path string, args ...interface{}) error {

	t1 := time.Now()

	defer func(){
		t2 := time.Since(t1)
		i.Logger.Status("time to index directory '%s' %v", path, t2)
	}()

	i.increment()
	defer i.decrement()

	abs_path, err := filepath.Abs(path)

	if err != nil {
		return err
	}

	cb := func(path string, info os.FileInfo) error {

		if info.IsDir() {
			return nil
		}

		return i.process_path(path, args...)
	}

	c := crawl.NewCrawler(abs_path)
	return c.Crawl(cb)
}

func (i *Indexer) IndexGeoJSONFeatureCollection(path string, args ...interface{}) error {

	t1 := time.Now()

	defer func(){
		t2 := time.Since(t1)
		i.Logger.Status("time to index feature collection '%s' %v", path, t2)
	}()

	i.increment()
	defer i.decrement()

	fh, err := i.readerFromPath(path)

	if err != nil {
		return err
	}

	defer fh.Close()

	body, err := ioutil.ReadAll(fh)

	if err != nil {
		return err
	}

	type FC struct {
		Type     string
		Features []interface{}
	}

	var collection FC

	err = json.Unmarshal(body, &collection)

	if err != nil {
		return err
	}

	for _, f := range collection.Features {

		feature, err := json.Marshal(f)

		if err != nil {
			return err
		}

		fh := bytes.NewBuffer(feature)
		err = i.process(fh, path, args...)

		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Indexer) IndexGeoJSONLS(path string, args ...interface{}) error {

	t1 := time.Now()

	defer func(){
		t2 := time.Since(t1)
		i.Logger.Status("time to index geojson-ls '%s' %v", path, t2)
	}()

	i.increment()
	defer i.decrement()

	fh, err := i.readerFromPath(path)

	if err != nil {
		return err
	}

	defer fh.Close()

	// see this - we're using ReadLine because it's entirely possible
	// that the raw GeoJSON (LS) will be too long for bufio.Scanner
	// see also - https://golang.org/pkg/bufio/#Reader.ReadLine
	// (20170822/thisisaaronland)

	reader := bufio.NewReader(fh)
	raw := bytes.NewBuffer([]byte(""))

	for {
		fragment, is_prefix, err := reader.ReadLine()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		raw.Write(fragment)

		if is_prefix {
			continue
		}

		fh := bytes.NewReader(raw.Bytes())

		err = i.process(fh, path, args...)

		if err != nil {
			return err
		}

		raw.Reset()
	}

	return nil
}

func (i *Indexer) IndexMetaFile(path string, data_root string, args ...interface{}) error {

	t1 := time.Now()

	defer func(){
		t2 := time.Since(t1)
		i.Logger.Status("time to index meta file '%s' %v", path, t2)
	}()

	i.increment()
	defer i.decrement()

	fh, err := i.readerFromPath(path)

	if err != nil {
		return err
	}

	defer fh.Close()

	reader, err := csv.NewDictReader(fh)

	if err != nil {
		return err
	}

	for {
		row, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		rel_path, ok := row["path"]

		if !ok {
			return errors.New("Missing path key")
		}

		// TO DO: make this work with a row["repo"] key
		// (20170809/thisisaaronland)

		file_path := filepath.Join(data_root, rel_path)

		err = i.process_path(file_path, args...)

		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Indexer) IndexFileList(path string, args ...interface{}) error {

	t1 := time.Now()

	defer func(){
		t2 := time.Since(t1)
		i.Logger.Status("time to index file list '%s' %v", path, t2)
	}()

	i.increment()
	defer i.decrement()

	fh, err := i.readerFromPath(path)

	if err != nil {
		return err
	}

	defer fh.Close()

	scanner := bufio.NewScanner(fh)

	for scanner.Scan() {

		file_path := scanner.Text()

		err = i.process_path(file_path, args...)

		if err != nil {
			return err
		}
	}

	err = scanner.Err()

	if err != nil {
		return err
	}

	return nil
}

func (i *Indexer) IsIndexing() bool {

	if atomic.LoadInt64(&i.count) > 0 {
		return true
	}

	return false
}

func (i *Indexer) readerFromPath(abs_path string) (io.ReadCloser, error) {

	if abs_path == STDIN {
		return os.Stdin, nil
	}

	fh, err := os.Open(abs_path)

	if err != nil {
		return nil, err
	}

	return fh, nil
}

func (i *Indexer) process_path(path string, args ...interface{}) error {

	abs_path, err := filepath.Abs(path)

	if err != nil {
		return err
	}

	fh, err := os.Open(abs_path)

	if err != nil {
		return err
	}

	defer fh.Close()

	return i.process(fh, abs_path, args...)
}

func (i *Indexer) process(fh io.Reader, path string, args ...interface{}) error {

     	t1 := time.Now()

	defer func(){
		t2 := time.Since(t1)
		i.Logger.Debug("time to process record '%s' %v", path, t2)		
	}()
	
	i.increment()
	defer i.decrement()

	ctx, err := ContextForPath(path)

	if err != nil {
		return err
	}

	err = i.Func(fh, ctx, args...)

	if err != nil {
		return err
	}

	atomic.AddInt64(&i.Indexed, 1)
	return nil
}

func (i *Indexer) increment() {
	atomic.AddInt64(&i.count, 1)
}

func (i *Indexer) decrement() {
	atomic.AddInt64(&i.count, -1)
}
