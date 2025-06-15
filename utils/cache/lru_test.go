package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLRUCache(t *testing.T) {
	m := NewLRUCache[string, int](3)

	m.Store("target-data", 1)

	m.Store("data-1", 1)

	m.Store("data-2", 1)

	// Access the target data to keep it in the cache
	m.Load("target-data")
	m.Store("data-3", 1)

	m.Store("data-4", 1)

	val, ok := m.Load("target-data")
	assert.True(t, ok)
	assert.Equal(t, 1, val)
}

func TestLRUCache_Expire(t *testing.T) {
	m := NewLRUCache[string, int](3)

	m.Store("target-data", 1)

	m.Store("data-1", 1)
	m.Store("data-2", 1)
	m.Store("data-3", 1)
	m.Store("data-4", 1)
	m.Store("data-5", 1)

	_, ok := m.Load("target-data")
	assert.False(t, ok)
}
