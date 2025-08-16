package result

import (
	"container/list"
	"sync"
	"time"
)

// BoundedCache implements an LRU cache with a maximum capacity
type BoundedCache[K comparable, V any] struct {
	mu       sync.RWMutex
	capacity int
	items    map[K]*list.Element
	order    *list.List
}

type boundedCacheItem[K comparable, V any] struct {
	key        K
	value      V
	expiration time.Time
}

// NewBoundedCache creates a new bounded cache with the specified capacity
func NewBoundedCache[K comparable, V any](capacity int) *BoundedCache[K, V] {
	return &BoundedCache[K, V]{
		capacity: capacity,
		items:    make(map[K]*list.Element),
		order:    list.New(),
	}
}

// Set adds or updates an item in the cache with a default TTL
func (c *BoundedCache[K, V]) Set(key K, value V) {
	c.SetT(key, value, time.Hour) // Default TTL of 1 hour
}

// SetT adds or updates an item in the cache with a specific TTL
func (c *BoundedCache[K, V]) SetT(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiration := time.Now().Add(ttl)
	item := &boundedCacheItem[K, V]{
		key:        key,
		value:      value,
		expiration: expiration,
	}

	// If key already exists, update it and move to front
	if elem, exists := c.items[key]; exists {
		elem.Value = item
		c.order.MoveToFront(elem)
		return
	}

	// If at capacity, remove oldest item
	if len(c.items) >= c.capacity {
		c.evictOldest()
	}

	// Add new item to front
	elem := c.order.PushFront(item)
	c.items[key] = elem

	// Set up expiration cleanup
	go func() {
		<-time.After(ttl)
		c.Delete(key)
	}()
}

// Get retrieves an item from the cache and marks it as recently used
func (c *BoundedCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var zero V
	elem, exists := c.items[key]
	if !exists {
		return zero, false
	}

	item := elem.Value.(*boundedCacheItem[K, V])

	// Check if expired
	if time.Now().After(item.expiration) {
		c.delete(key)
		return zero, false
	}

	// Move to front (mark as recently used)
	c.order.MoveToFront(elem)
	return item.value, true
}

// Has checks if a key exists in the cache without updating access time
func (c *BoundedCache[K, V]) Has(key K) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	elem, exists := c.items[key]
	if !exists {
		return false
	}

	item := elem.Value.(*boundedCacheItem[K, V])
	if time.Now().After(item.expiration) {
		return false
	}

	return true
}

// Delete removes an item from the cache
func (c *BoundedCache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.delete(key)
}

// delete removes an item from the cache (internal, assumes lock is held)
func (c *BoundedCache[K, V]) delete(key K) {
	if elem, exists := c.items[key]; exists {
		c.order.Remove(elem)
		delete(c.items, key)
	}
}

// Clear removes all items from the cache
func (c *BoundedCache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[K]*list.Element)
	c.order.Init()
}

// GetOrSet retrieves an item or creates it if it doesn't exist
func (c *BoundedCache[K, V]) GetOrSet(key K, createFunc func() (V, error)) (V, error) {
	// Try to get the value first
	value, ok := c.Get(key)
	if ok {
		return value, nil
	}

	// Create new value
	newValue, err := createFunc()
	if err != nil {
		return newValue, err
	}

	// Set the new value
	c.Set(key, newValue)
	return newValue, nil
}

// Size returns the current number of items in the cache
func (c *BoundedCache[K, V]) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Capacity returns the maximum capacity of the cache
func (c *BoundedCache[K, V]) Capacity() int {
	return c.capacity
}

// evictOldest removes the least recently used item (assumes lock is held)
func (c *BoundedCache[K, V]) evictOldest() {
	if c.order.Len() == 0 {
		return
	}

	elem := c.order.Back()
	if elem != nil {
		item := elem.Value.(*boundedCacheItem[K, V])
		c.delete(item.key)
	}
}

// Range iterates over all items in the cache
func (c *BoundedCache[K, V]) Range(callback func(key K, value V) bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for elem := c.order.Front(); elem != nil; elem = elem.Next() {
		item := elem.Value.(*boundedCacheItem[K, V])

		// Skip expired items
		if time.Now().After(item.expiration) {
			continue
		}

		if !callback(item.key, item.value) {
			break
		}
	}
}
