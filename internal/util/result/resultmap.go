package result

import (
	"seanime/internal/util"
)

type Map[K interface{}, V any] struct {
	store util.RWMutexMap
}

type mapItem[K interface{}, V any] struct {
	value V
}

func NewResultMap[K interface{}, V any]() *Map[K, V] {
	return &Map[K, V]{}
}

func (c *Map[K, V]) Set(key K, value V) {
	c.store.Store(key, &mapItem[K, V]{value})
}

func (c *Map[K, V]) Get(key K) (V, bool) {
	item, ok := c.store.Load(key)
	if !ok {
		return (&mapItem[K, V]{}).value, false
	}
	ci := item.(*mapItem[K, V])
	return ci.value, true
}

func (c *Map[K, V]) Has(key K) bool {
	_, ok := c.store.Load(key)
	return ok
}

func (c *Map[K, V]) GetOrSet(key K, createFunc func() (V, error)) (V, error) {
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

func (c *Map[K, V]) Delete(key K) {
	c.store.Delete(key)
}

func (c *Map[K, V]) Clear() {
	c.store.Range(func(key interface{}, value interface{}) bool {
		c.store.Delete(key)
		return true
	})
}

// ClearN clears the map and returns the number of items cleared
func (c *Map[K, V]) ClearN() int {
	count := 0
	c.store.Range(func(key interface{}, value interface{}) bool {
		c.store.Delete(key)
		count++
		return true
	})
	return count
}

func (c *Map[K, V]) Range(callback func(key K, value V) bool) {
	c.store.Range(func(key, value interface{}) bool {
		ci := value.(*mapItem[K, V])
		return callback(key.(K), ci.value)
	})
}

func (c *Map[K, V]) Values() []V {
	values := make([]V, 0)
	c.store.Range(func(key, value interface{}) bool {
		item := value.(*mapItem[K, V]) // Correct type assertion
		values = append(values, item.value)
		return true
	})
	return values
}
