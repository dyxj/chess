package store

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMemCache(t *testing.T) {
	cache := NewMemCache()

	assert.NotNil(t, cache)
	assert.NotNil(t, cache.items)
	assert.Equal(t, 0, len(cache.items))
}

func TestMemCache_Add(t *testing.T) {
	cache := NewMemCache()

	t.Run("successful add", func(t *testing.T) {
		err := cache.Add("key1", "value1", time.Time{})
		assert.NoError(t, err)

		value, exists := cache.Find("key1")
		assert.True(t, exists)
		assert.Equal(t, "value1", value)
	})

	t.Run("add with expiry", func(t *testing.T) {
		expiry := time.Now().Add(time.Hour)
		err := cache.Add("key2", "value2", expiry)
		assert.NoError(t, err)

		value, exists := cache.Find("key2")
		assert.True(t, exists)
		assert.Equal(t, "value2", value)
	})

	t.Run("add duplicate key returns error", func(t *testing.T) {
		cache.Add("duplicate", "first", time.Time{})

		err := cache.Add("duplicate", "second", time.Time{})
		assert.Equal(t, ErrKeyAlreadyExists, err)

		value, exists := cache.Find("duplicate")
		assert.True(t, exists)
		assert.Equal(t, "first", value)
	})
}

func TestMemCache_Find(t *testing.T) {
	cache := NewMemCache()

	t.Run("find existing key", func(t *testing.T) {
		cache.Add("existing", "value", time.Time{})

		value, exists := cache.Find("existing")
		assert.True(t, exists)
		assert.Equal(t, "value", value)
	})

	t.Run("find non-existing key", func(t *testing.T) {
		value, exists := cache.Find("non-existing")
		assert.False(t, exists)
		assert.Nil(t, value)
	})

	t.Run("find expired key", func(t *testing.T) {
		pastTime := time.Now().Add(-time.Hour)
		cache.Add("expired", "value", pastTime)

		value, exists := cache.Find("expired")
		assert.False(t, exists)
		assert.Nil(t, value)
	})

	t.Run("find key that expires in future", func(t *testing.T) {
		futureTime := time.Now().Add(time.Hour)
		cache.Add("future", "value", futureTime)

		value, exists := cache.Find("future")
		assert.True(t, exists)
		assert.Equal(t, "value", value)
	})
}

func TestMemCache_Delete(t *testing.T) {
	cache := NewMemCache()

	t.Run("delete existing key", func(t *testing.T) {
		cache.Add("to-delete", "value", time.Time{})

		_, exists := cache.Find("to-delete")
		assert.True(t, exists)

		cache.Delete("to-delete")

		_, exists = cache.Find("to-delete")
		assert.False(t, exists)
	})

	t.Run("delete non-existing key doesn't panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			cache.Delete("non-existing")
		})
	})
}

func TestMemCache_Update(t *testing.T) {
	cache := NewMemCache()

	t.Run("update existing key", func(t *testing.T) {
		cache.Add("update-me", "original", time.Time{})

		newExpiry := time.Now().Add(time.Hour)
		err := cache.Update("update-me", "updated", newExpiry)
		assert.NoError(t, err)

		value, exists := cache.Find("update-me")
		assert.True(t, exists)
		assert.Equal(t, "updated", value)
	})

	t.Run("update non-existing key returns error", func(t *testing.T) {
		err := cache.Update("non-existing", "value", time.Time{})
		assert.Equal(t, ErrKeyNotFound, err)
	})

	t.Run("update expired key returns error", func(t *testing.T) {
		pastTime := time.Now().Add(-time.Hour)
		cache.Add("expired-update", "value", pastTime)

		err := cache.Update("expired-update", "new-value", time.Time{})
		assert.Equal(t, ErrKeyNotFound, err)
	})
}

func TestMemCache_Set(t *testing.T) {
	cache := NewMemCache()

	t.Run("set new key", func(t *testing.T) {
		err := cache.Set("new-key", "new-value", time.Time{})
		assert.NoError(t, err)

		value, exists := cache.Find("new-key")
		assert.True(t, exists)
		assert.Equal(t, "new-value", value)
	})

	t.Run("set existing key updates value", func(t *testing.T) {
		cache.Add("existing-set", "original", time.Time{})

		err := cache.Set("existing-set", "updated", time.Time{})
		assert.NoError(t, err)

		value, exists := cache.Find("existing-set")
		assert.True(t, exists)
		assert.Equal(t, "updated", value)
	})

	t.Run("set with expiry", func(t *testing.T) {
		expiry := time.Now().Add(time.Hour)
		err := cache.Set("with-expiry", "value", expiry)
		assert.NoError(t, err)

		value, exists := cache.Find("with-expiry")
		assert.True(t, exists)
		assert.Equal(t, "value", value)
	})
}

func TestMemCache_ExpiryHandling(t *testing.T) {
	cache := NewMemCache()

	t.Run("item expires correctly", func(t *testing.T) {
		expiry := time.Now().Add(50 * time.Millisecond)
		cache.Add("short-lived", "value", expiry)

		value, exists := cache.Find("short-lived")
		assert.True(t, exists)
		assert.Equal(t, "value", value)

		<-time.After(100 * time.Millisecond)

		value, exists = cache.Find("short-lived")
		assert.False(t, exists)
		assert.Nil(t, value)
	})

	t.Run("zero expiry means no expiration", func(t *testing.T) {
		cache.Add("no-expiry", "value", time.Time{})

		<-time.After(10 * time.Millisecond)
		value, exists := cache.Find("no-expiry")
		assert.True(t, exists)
		assert.Equal(t, "value", value)
	})
}

func TestMemCache_MaintenanceJobs(t *testing.T) {
	cache := NewMemCache()
	stop := make(chan struct{})

	t.Run("cleanup job removes expired items", func(t *testing.T) {
		shortExpiry := time.Now().Add(50 * time.Millisecond)
		cache.Add("expire1", "value1", shortExpiry)
		cache.Add("expire2", "value2", shortExpiry)
		cache.Add("permanent", "value3", time.Time{})

		cache.StartMaintenanceJobs(10, 200*time.Millisecond, 100*time.Millisecond, stop)

		<-time.After(200 * time.Millisecond)

		cache.mu.RLock()
		_, exists1 := cache.items["expire1"]
		_, exists2 := cache.items["expire2"]
		_, exists3 := cache.items["permanent"]
		cache.mu.RUnlock()

		assert.False(t, exists1)
		assert.False(t, exists2)
		assert.True(t, exists3)
	})

	t.Run("maintenance jobs can be stopped", func(t *testing.T) {
		close(stop)
		<-time.After(50 * time.Millisecond)
	})

	t.Run("shrink job works correctly", func(t *testing.T) {
		cache := NewMemCache()
		stop := make(chan struct{})
		defer close(stop)

		for i := 0; i < 15; i++ {
			cache.Add(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i), time.Time{})
		}

		cache.StartMaintenanceJobs(5, 50*time.Millisecond, time.Hour, stop)

		<-time.After(100 * time.Millisecond)

		for i := 0; i < 15; i++ {
			value, exists := cache.Find(fmt.Sprintf("key%d", i))
			assert.True(t, exists)
			assert.Equal(t, fmt.Sprintf("value%d", i), value)
		}
	})
}

func TestMemCache_ConcurrencySafety(t *testing.T) {
	cache := NewMemCache()

	t.Run("concurrent reads and writes", func(t *testing.T) {
		const numGoroutines = 10
		const numOperations = 100

		var wg sync.WaitGroup
		wg.Add(numGoroutines * 3)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numOperations; j++ {
					key := fmt.Sprintf("key_%d_%d", id, j%10)
					cache.Find(key)
				}
			}(i)
		}

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numOperations; j++ {
					key := fmt.Sprintf("key_%d_%d", id, j%10)
					value := fmt.Sprintf("value_%d_%d", id, j)
					cache.Set(key, value, time.Time{})
				}
			}(i)
		}

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numOperations; j++ {
					key := fmt.Sprintf("key_%d_%d", id, j%10)
					cache.Delete(key)
				}
			}(i)
		}

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatal("Test timed out - possible deadlock")
		}
	})
}

func TestMemCache_EdgeCases(t *testing.T) {
	cache := NewMemCache()

	t.Run("nil values are supported", func(t *testing.T) {
		err := cache.Add("nil-value", nil, time.Time{})
		assert.NoError(t, err)

		value, exists := cache.Find("nil-value")
		assert.True(t, exists)
		assert.Nil(t, value)
	})

	t.Run("empty string keys are supported", func(t *testing.T) {
		err := cache.Add("", "empty-key-value", time.Time{})
		assert.NoError(t, err)

		value, exists := cache.Find("")
		assert.True(t, exists)
		assert.Equal(t, "empty-key-value", value)
	})

	t.Run("complex data types", func(t *testing.T) {
		complexValue := map[string]interface{}{
			"nested": map[string]int{
				"count": 42,
			},
			"list": []string{"a", "b", "c"},
		}

		err := cache.Add("complex", complexValue, time.Time{})
		assert.NoError(t, err)

		value, exists := cache.Find("complex")
		assert.True(t, exists)
		assert.Equal(t, complexValue, value)
	})
}

func TestMemCache_deleteIfExpired(t *testing.T) {
	cache := NewMemCache()

	t.Run("non-expired item returns false", func(t *testing.T) {
		item := cacheItem{
			value:  "test",
			expiry: time.Now().Add(time.Hour),
		}

		deleted := cache.deleteIfExpired("test-key", item)
		assert.False(t, deleted)
	})

	t.Run("expired item returns true", func(t *testing.T) {
		cache.items["test-key"] = cacheItem{
			value:  "test",
			expiry: time.Now().Add(-time.Hour),
		}

		item := cache.items["test-key"]
		deleted := cache.deleteIfExpired("test-key", item)
		assert.True(t, deleted)

		_, exists := cache.items["test-key"]
		assert.False(t, exists)
	})

	t.Run("zero expiry returns false", func(t *testing.T) {
		item := cacheItem{
			value:  "test",
			expiry: time.Time{},
		}

		deleted := cache.deleteIfExpired("test-key", item)
		assert.False(t, deleted)
	})
}

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
