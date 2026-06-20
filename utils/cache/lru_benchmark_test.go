package cache

import (
	"sync"
	"testing"
)

var MAX_SIZE = 1000

func BenchmarkLRUCacheStore(b *testing.B) {
	cache := NewLRUCache[string, string](MAX_SIZE)
	for i := 0; i < b.N; i++ {
		key := string(rune(i))
		cache.Store(key, "test value")
	}
}

func BenchmarkLRUCacheLoad(b *testing.B) {
	cache := NewLRUCache[string, string](MAX_SIZE)
	for i := 0; i < b.N; i++ {
		key := string(rune(i))
		cache.Store(key, "test value")
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		key := string(rune(i))
		cache.Load(key)
	}
	b.StopTimer()
}

/**
 * Benchmark for normal map for comparison
 */

func BenchmarkNormalMapStore(b *testing.B) {
	m := sync.Map{}
	for i := 0; i < b.N; i++ {
		key := string(rune(i))
		m.Store(key, "test value")
	}
}

func BenchmarkNormalMapLoad(b *testing.B) {
	m := sync.Map{}
	for i := 0; i < b.N; i++ {
		key := string(rune(i))
		m.Store(key, "test value")
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		key := string(rune(i))
		m.Load(key)
	}
	b.StopTimer()
}
