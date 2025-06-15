package cache

import (
	"sync"
)

type LRUCache[K comparable, V any] struct {
	current     *sync.Map
	prev        *sync.Map
	lock        sync.Mutex
	currentSize int
	// The maximum size guaranteed to be stored in the cache
	maxSize int
}

func NewLRUCache[K comparable, V any](maxSize int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		current:     &sync.Map{},
		prev:        &sync.Map{},
		lock:        sync.Mutex{},
		currentSize: 0,
		maxSize:     maxSize,
	}
}

func (m *LRUCache[K, V]) Load(key K) (value V, ok bool) {
	_value, ok := m.current.Load(key)
	if ok {
		return _value.(V), ok
	}
	_value, ok = m.prev.Load(key)
	if ok {
		m.Store(key, _value.(V)) // Restore to current
		return _value.(V), ok
	}
	return
}

func (m *LRUCache[K, V]) Store(key K, value V) {
	m.current.Store(key, value)

	m.lock.Lock()
	defer m.lock.Unlock()
	m.currentSize++

	// Move current to prev if current size exceeds maxSize
	if m.currentSize >= m.maxSize {
		m.current, m.prev = &sync.Map{}, m.current
		m.currentSize = 0
	}
}
