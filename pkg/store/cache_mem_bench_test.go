package store

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkMemCache_Add(b *testing.B) {
	cache := NewMemCache()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Add(key, "value", time.Time{})
	}
}

func BenchmarkMemCache_Find(b *testing.B) {
	cache := NewMemCache()

	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Add(key, "value", time.Time{})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%1000)
		cache.Find(key)
	}
}

func BenchmarkMemCache_Set(b *testing.B) {
	cache := NewMemCache()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, "value", time.Time{})
	}
}

func BenchmarkMemCache_ConcurrentAccess(b *testing.B) {
	cache := NewMemCache()

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Add(key, "value", time.Time{})
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key%d", i%100)
			if i%2 == 0 {
				cache.Find(key)
			} else {
				cache.Set(key, "new-value", time.Time{})
			}
			i++
		}
	})
}
