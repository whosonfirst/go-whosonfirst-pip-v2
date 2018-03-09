package cache

import (
	"errors"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
)

type FSCache struct {
	Cache
	Logger     *log.WOFLogger
	data_root  string
	repo_index []string       // this is a list of repo names
	repo_map   map[string]int // this is a list of WOF ID -> index of repo name in `repo_index`
	mu         *sync.RWMutex
	hits       int64
	misses     int64
	evictions  int64
	keys       int64
}

func NewFSCache(data_root string) (Cache, error) {

	_, err := os.Stat(data_root)

	if os.IsNotExist(err) {
		return nil, err
	}

	logger := log.SimpleWOFLogger("source")

	i := make([]string, 0)
	m := make(map[string]int)

	mu := new(sync.RWMutex)

	c := FSCache{
		Logger:     logger,
		data_root:  data_root,
		repo_map:   m,
		repo_index: i,
		mu:         mu,
		hits:       int64(0),
		misses:     int64(0),
		evictions:  int64(0),
		keys:       int64(0),
	}

	return &c, nil
}

func (c *FSCache) Close() error {
	return nil
}

func (c *FSCache) Get(key string) (CacheItem, error) {

	// to do: timings that don't slow everything down the way
	// go-whosonfirst-timer does now (20170915/thisisaaronland)

	c.Logger.Info("GET %s", key)

	c.mu.RLock()

	idx, ok := c.repo_map[key]
	repo := c.repo_index[idx]

	c.mu.RUnlock()

	if !ok {
		return nil, errors.New("Unable to determine repo for ID")
	}

	abs_path, err := c.str_id2abspath(key, repo)

	if err != nil {
		return nil, err
	}

	f, err := feature.LoadWOFFeatureFromFile(abs_path)

	if err != nil {
		atomic.AddInt64(&c.misses, 1)
		return nil, err
	}

	fc, err := NewFeatureCache(f)

	if err != nil {
		return nil, err
	}

	atomic.AddInt64(&c.hits, 1)
	return fc, nil
}

func (c *FSCache) Set(key string, i CacheItem) error {

	c.Logger.Info("SET %s", key)

	s := i.SPR()
	repo := s.Repo()

	if repo == "" {
		return errors.New("Unable to determine wof:repo for feature")
	}

	_, err := c.str_id2abspath(key, repo)

	if err != nil {
		return err
	}

	c.mu.Lock()

	idx := -1

	for i, name := range c.repo_index {
		if name == repo {
			idx = i
			break
		}
	}

	if idx == -1 {

		c.repo_index = append(c.repo_index, repo)
		idx = len(c.repo_index) - 1
	}

	c.repo_map[key] = idx

	c.mu.Unlock()
	return nil
}

func (c *FSCache) Size() int64 {
	return atomic.LoadInt64(&c.keys)
}

func (c *FSCache) Hits() int64 {
	return atomic.LoadInt64(&c.hits)
}

func (c *FSCache) Misses() int64 {
	return atomic.LoadInt64(&c.misses)
}

func (c *FSCache) Evictions() int64 {
	return atomic.LoadInt64(&c.evictions)
}

func (c *FSCache) str_id2abspath(key string, repo string) (string, error) {

	repo_path := filepath.Join(c.data_root, repo)

	_, err := os.Stat(repo_path)

	if os.IsNotExist(err) {
		return "", err
	}

	data_path := filepath.Join(repo_path, "data")

	_, err = os.Stat(data_path)

	if os.IsNotExist(err) {
		return "", err
	}

	wofid, err := strconv.ParseInt(key, 10, 64)

	if err != nil {
		return "", err
	}

	abs_path, err := uri.Id2AbsPath(data_path, wofid)

	if os.IsNotExist(err) {
		return "", err
	}

	_, err = os.Stat(abs_path)

	if os.IsNotExist(err) {
		return "", err
	}

	return abs_path, nil
}
