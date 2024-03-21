package filecache

import (
	"github.com/goccy/go-json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CacheStore represents a single-process, file-based, key/value cache store.
type CacheStore struct {
	filePath string
	mu       sync.Mutex
	data     map[string]*cacheItem
}

// Bucket represents a cache bucket with a name and TTL.
type Bucket struct {
	name string
	ttl  time.Duration
}

func NewBucket(name string, ttl time.Duration) Bucket {
	return Bucket{name: name, ttl: ttl}
}

// Cacher represents a single-process, file-based, key/value cache.
type Cacher struct {
	dir    string
	stores map[string]*CacheStore
	mu     sync.Mutex
}

type cacheItem struct {
	Value      interface{} `json:"value"`
	Expiration time.Time   `json:"expiration"`
}

// NewCacher creates a new instance of Cacher.
func NewCacher(dir string) (*Cacher, error) {
	// Check if the directory exists
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return &Cacher{
		stores: make(map[string]*CacheStore),
		dir:    dir,
	}, nil
}

// Close closes all the cache stores.
func (c *Cacher) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, store := range c.stores {
		if err := store.saveToFile(); err != nil {
			return err
		}
	}
	return nil
}

// getStore returns a cache store for the given bucket name and TTL.
func (c *Cacher) getStore(name string) (*CacheStore, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	store, ok := c.stores[name]
	if !ok {
		store = &CacheStore{
			filePath: filepath.Join(c.dir, name+".cache"),
			data:     make(map[string]*cacheItem),
		}
		if err := store.loadFromFile(); err != nil {
			return nil, err
		}
		c.stores[name] = store
	}
	return store, nil
}

// Set sets the value for the given key in the given bucket.
func (c *Cacher) Set(bucket Bucket, key string, value interface{}) error {
	store, err := c.getStore(bucket.name)
	if err != nil {
		return err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	store.data[key] = &cacheItem{Value: value, Expiration: time.Now().Add(bucket.ttl)}
	return store.saveToFile()
}

// Get retrieves the value for the given key from the given bucket.
func (c *Cacher) Get(bucket Bucket, key string, out interface{}) (bool, error) {
	store, err := c.getStore(bucket.name)
	if err != nil {
		return false, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	item, ok := store.data[key]
	if !ok {
		return false, nil
	}
	if time.Now().After(item.Expiration) {
		delete(store.data, key)
		_ = store.saveToFile() // Ignore errors here
		return false, nil
	}
	data, err := json.Marshal(item.Value)
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal(data, out)
}

// Delete deletes the value for the given key from the given bucket.
func (c *Cacher) Delete(bucket Bucket, key string) error {
	store, err := c.getStore(bucket.name)
	if err != nil {
		return err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	delete(store.data, key)
	return store.saveToFile()
}

func (c *Cacher) DeleteAll(bucket Bucket) error {
	store, err := c.getStore(bucket.name)
	if err != nil {
		return err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	store.data = make(map[string]*cacheItem)
	return store.saveToFile()
}

func (cs *CacheStore) loadFromFile() error {
	file, err := os.Open(cs.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File does not exist, so nothing to load
		}
		return err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&cs.data); err != nil {
		return err
	}
	return nil
}

func (cs *CacheStore) saveToFile() error {
	file, err := os.Create(cs.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(cs.data); err != nil {
		return err
	}
	return nil
}
