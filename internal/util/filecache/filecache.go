package filecache

import (
	"github.com/goccy/go-json"
	"os"
	"path/filepath"
	"strings"
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

// Empty empties the given bucket.
func (c *Cacher) Empty(bucket Bucket) error {
	store, err := c.getStore(bucket.name)
	if err != nil {
		return err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	store.data = make(map[string]*cacheItem)
	return store.saveToFile()
}

// Remove removes the given bucket.
func (c *Cacher) Remove(bucketName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.stores[bucketName]; ok {
		delete(c.stores, bucketName)
	}

	_ = os.Remove(filepath.Join(c.dir, bucketName+".cache"))

	return nil
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// RemoveAllBy removes all files in the cache directory that match the given filter.
func (c *Cacher) RemoveAllBy(filter func(filename string) bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := filepath.Walk(c.dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if !strings.HasSuffix(info.Name(), ".cache") {
				return nil
			}
			if filter(info.Name()) {
				if err := os.Remove(filepath.Join(c.dir, info.Name())); err != nil {
					return err
				}
			}
		}
		return nil
	})

	c.stores = make(map[string]*CacheStore)
	return err
}

// ClearMediastreamVideoFiles clears all mediastream video file caches.
func (c *Cacher) ClearMediastreamVideoFiles() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove the contents of the directory
	files, err := os.ReadDir(filepath.Join(c.dir, "videofiles"))
	if err != nil {
		return nil
	}
	for _, file := range files {
		_ = os.RemoveAll(filepath.Join(c.dir, "videofiles", file.Name()))
	}

	c.stores = make(map[string]*CacheStore)
	return err
}

// TrimMediastreamVideoFiles clears all mediastream video file caches if the number of files exceeds the given limit.
func (c *Cacher) TrimMediastreamVideoFiles() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove the contents of the directory
	files, err := os.ReadDir(filepath.Join(c.dir, "videofiles"))
	if err != nil {
		return nil
	}

	if len(files) > 3 {
		for _, file := range files {
			_ = os.RemoveAll(filepath.Join(c.dir, "videofiles", file.Name()))
		}
	}

	c.stores = make(map[string]*CacheStore)
	return err
}

func (c *Cacher) GetMediastreamVideoFilesTotalSize() (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := os.Stat(filepath.Join(c.dir, "videofiles"))
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	var totalSize int64
	err = filepath.Walk(filepath.Join(c.dir, "videofiles"), func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return totalSize, nil
}

// GetTotalSize returns the total size of all files in the cache directory that match the given filter.
// The size is in bytes.
func (c *Cacher) GetTotalSize() (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var totalSize int64
	err := filepath.Walk(c.dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return totalSize, nil
}
