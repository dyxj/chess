package store

import (
	"sync"
	"time"

	"github.com/dyxj/chess/pkg/safe"
)

type MemCache struct {
	mu sync.RWMutex
	// key: value
	items map[string]cacheItem

	shrinkJobOnce  sync.Once
	cleanupJobOnce sync.Once

	downsizeThreshold     int
	downsizeCheckInterval time.Duration
	cleanupInterval       time.Duration
}

func NewMemCache() *MemCache {
	return &MemCache{
		items: make(map[string]cacheItem, 20),
	}
}

type cacheItem struct {
	value  any
	expiry time.Time // 0 for no expiry
}

// Add key value pair to cache with expiry.
// If expiry is zero, the item will never expire.
// If the key already exists, it returns ErrKeyAlreadyExists.
func (c *MemCache) Add(key string, value any, expiry time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.findItem(key)
	if ok {
		return ErrKeyAlreadyExists
	}
	c.items[key] = cacheItem{
		value:  value,
		expiry: expiry,
	}
	return nil
}

// Find returns nil and false if it doesn't exist
func (c *MemCache) Find(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.findItem(key)
	if !ok {
		return nil, false
	}

	return item.value, ok
}

func (c *MemCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.delete(key)
}

// Update key value pair in cache with expiry.
// If expiry is zero, the item will never expire.
// If the key doesn't exist, it returns ErrKeyNotFound.
func (c *MemCache) Update(key string, value any, expiry time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.findItem(key)
	if !ok {
		return ErrKeyNotFound
	}
	c.items[key] = cacheItem{
		value:  value,
		expiry: expiry,
	}
	return nil
}

// Set key value pair in cache with expiry.
// If expiry is zero, the item will never expire.
// If the key doesn't exist, it will be created.
// If the key already exists, it will be updated.
func (c *MemCache) Set(key string, value any, expiry time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = cacheItem{
		value:  value,
		expiry: expiry,
	}
	return nil
}

func (c *MemCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]cacheItem, 20)
}

func (c *MemCache) findItem(key string) (cacheItem, bool) {
	item, ok := c.items[key]
	if !ok {
		return cacheItem{}, false
	}
	if c.deleteIfExpired(key, item) {
		return cacheItem{}, false
	}

	return item, ok
}

func (c *MemCache) StartMaintenanceJobs(
	downsizeThreshold int,
	downsizeCheckInterval time.Duration,
	cleanupInterval time.Duration,
	stop <-chan struct{},
) {
	c.cleanupJobOnce.Do(func() {
		ticker := time.NewTicker(cleanupInterval)
		safe.Go(
			func() {
				for {
					select {
					case <-ticker.C:
						c.cleanupJob()
					case <-stop:
						ticker.Stop()
						return
					}
				}
			},
		)
	})

	c.shrinkJobOnce.Do(func() {
		ticker := time.NewTicker(downsizeCheckInterval)
		safe.Go(func() {
			for {
				select {
				case <-ticker.C:
					c.shrinkIfRequired(downsizeThreshold)
				case <-stop:
					ticker.Stop()
					return
				}
			}
		})
	})
}

func (c *MemCache) cleanupJob() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, item := range c.items {
		c.deleteIfExpired(key, item)
	}
}

func (c *MemCache) shrinkIfRequired(downsizeThreshold int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.items) > downsizeThreshold {
		c.shrink()
	}
}

func (c *MemCache) shrink() {
	newRooms := make(map[string]cacheItem, len(c.items))

	for code, room := range c.items {
		newRooms[code] = room
	}

	c.items = newRooms
}

func (c *MemCache) delete(key string) {
	delete(c.items, key)
}

func (c *MemCache) deleteIfExpired(key string, item cacheItem) bool {
	if item.expiry.IsZero() {
		return false
	}
	if item.expiry.Before(time.Now()) {
		c.delete(key)
		return true
	}
	return false
}
