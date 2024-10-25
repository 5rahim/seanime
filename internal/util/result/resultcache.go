package result

import (
	"seanime/internal/constants"
	"seanime/internal/util"
	"time"
)

type Cache[K interface{}, V any] struct {
	store util.RWMutexMap
}

type cacheItem[K interface{}, V any] struct {
	value      V
	expiration time.Time
}

func NewCache[K interface{}, V any]() *Cache[K, V] {
	return &Cache[K, V]{}
}

func (c *Cache[K, V]) Set(key K, value V) {
	ttl := constants.GcTime
	c.store.Store(key, &cacheItem[K, V]{value, time.Now().Add(ttl)})
	go func() {
		<-time.After(ttl)
		c.Delete(key)
	}()
}

func (c *Cache[K, V]) SetT(key K, value V, ttl time.Duration) {
	c.store.Store(key, &cacheItem[K, V]{value, time.Now().Add(ttl)})
	go func() {
		<-time.After(ttl)
		c.Delete(key)
	}()
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	item, ok := c.store.Load(key)
	if !ok {
		return (&cacheItem[K, V]{}).value, false
	}
	ci := item.(*cacheItem[K, V])
	if time.Now().After(ci.expiration) {
		c.Delete(key)
		return (&cacheItem[K, V]{}).value, false
	}
	return ci.value, true
}

func (c *Cache[K, V]) Pop() (K, V, bool) {
	var key K
	var value V
	var ok bool
	c.store.Range(func(k, v interface{}) bool {
		key = k.(K)
		value = v.(*cacheItem[K, V]).value
		ok = true
		c.store.Delete(k)
		return false
	})
	return key, value, ok
}

func (c *Cache[K, V]) Has(key K) bool {
	_, ok := c.store.Load(key)
	return ok
}

func (c *Cache[K, V]) GetOrSet(key K, createFunc func() (V, error)) (V, error) {
	value, ok := c.Get(key)
	if ok {
		return value, nil
	}

	newValue, err := createFunc()
	if err != nil {
		return newValue, err
	}
	c.Set(key, newValue)
	return newValue, nil
}

func (c *Cache[K, V]) Delete(key K) {
	c.store.Delete(key)
}

func (c *Cache[K, V]) Clear() {
	c.store.Range(func(key interface{}, value interface{}) bool {
		c.store.Delete(key)
		return true
	})
}

func (c *Cache[K, V]) Range(callback func(key K, value V) bool) {
	c.store.Range(func(key, value interface{}) bool {
		ci := value.(*cacheItem[K, V])
		return callback(key.(K), ci.value)
	})
}
