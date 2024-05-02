package transcoder

import "sync"

type CMap[K comparable, V any] struct {
	data map[K]V
	lock sync.RWMutex
}

func NewCMap[K comparable, V any]() CMap[K, V] {
	return CMap[K, V]{
		data: make(map[K]V),
	}
}

func (m *CMap[K, V]) Get(key K) (V, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	ret, ok := m.data[key]
	return ret, ok
}

func (m *CMap[K, V]) GetOrCreate(key K, create func() V) (V, bool) {
	m.lock.RLock()
	ret, ok := m.data[key]
	if ok {
		m.lock.RUnlock()
		return ret, false
	}
	m.lock.RUnlock()

	m.lock.Lock()
	defer m.lock.Unlock()

	ret, ok = m.data[key]
	if ok {
		return ret, false
	}

	val := create()
	m.data[key] = val
	return val, true
}

func (m *CMap[K, V]) GetOrSet(key K, val V) (V, bool) {
	return m.GetOrCreate(key, func() V { return val })
}

func (m *CMap[K, V]) Set(key K, val V) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.data[key] = val
}

func (m *CMap[K, V]) Remove(key K) {
	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.data, key)
}

func (m *CMap[K, V]) GetAndRemove(key K) (V, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	val, ok := m.data[key]
	delete(m.data, key)
	return val, ok
}
