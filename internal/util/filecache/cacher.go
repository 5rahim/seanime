package filecache

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Cacher single-process, file-based, key/value cache.
type Cacher struct {
	dir    string
	stores map[string]*Store
	mu     sync.Mutex
}

type NewCacherOptions struct {
	Dir string
}

// NewCacher creates a new instance of Cacher.
// It can and should be a shared instance.
// Close() must be called after any usage of the Cacher.
//
// Example:
//
//	func method(ctx *AnyCtx) {
//	  cacher := ctx.GetCacher()
//	  defer cacher.Close()
//	  // ...
//	}
func NewCacher(opts *NewCacherOptions) *Cacher {
	// Check if the directory exists
	_, err := os.Stat(opts.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(opts.Dir, 0755); err != nil {

			}
		}
	}
	return &Cacher{
		stores: make(map[string]*Store),
		dir:    opts.Dir,
	}
}

// Close closes all the stores.
// It can be called multiple times.
func (c *Cacher) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, store := range c.stores {
		if err := store.Close(); err == nil {
			delete(c.stores, k)
		}
	}
	return nil
}

// Set sets the value for the given key in the given bucket.
// It can be called after Close() is called.
func (c *Cacher) Set(bucket Bucket, key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	store, err := c.getStore(bucket.name, bucket.ttl)
	if err != nil {
		return err
	}
	// Set add the bucket name to the key to avoid TTL conflicts.
	return store.Set(bucket.name+":"+key, value)
}

func (c *Cacher) Get(bucket Bucket, key string, out interface{}) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	store, err := c.getStore(bucket.name, bucket.ttl)
	if err != nil {
		return false, err
	}
	return store.Get(bucket.name+":"+key, out)
}

func (c *Cacher) Delete(bucket Bucket, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	store, err := c.getStore(bucket.name, bucket.ttl)
	if err != nil {
		return err
	}
	return store.Delete(bucket.name + ":" + key)
}

func (c *Cacher) getStore(bucket string, ttl time.Duration) (*Store, error) {
	store, found := c.stores[bucket]
	if !found {
		println("Creating new store for bucket: ", bucket, " with TTL: "+ttl.String())
		var err error
		opts := Options{
			TTL:        ttl,
			BucketName: bucket,
			Path:       filepath.Join(c.dir, "cache.db"),
		}
		store, err = NewStore(opts)
		if err != nil {
			return nil, err
		}
		c.stores[bucket] = store
	}
	return store, nil
}
