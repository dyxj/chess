package room

import (
	"sync"
	"time"
)

const peakSize = 10_000
const downsizeThreshold = peakSize / 10
const downsizeCheckInterval = 24 * time.Hour
const cleanupInterval = 12 * time.Hour

type MemCache struct {
	mu sync.RWMutex
	// code: room
	rooms map[string]Room

	shrinkJobOnce  sync.Once
	cleanupJobOnce sync.Once
}

func NewMemCache() *MemCache {
	return &MemCache{
		rooms: make(map[string]Room, 20),
	}
}

func (c *MemCache) Add(room Room) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	room, ok := c.rooms[room.Code]
	if ok {
		return ErrCodeAlreadyExists
	}
	c.rooms[room.Code] = room
	return nil
}

func (c *MemCache) Find(code string) (Room, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	room, ok := c.rooms[code]
	return room, ok
}

func (c *MemCache) Update(room Room) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.rooms[room.Code]
	if !ok {
		return ErrRoomNotFound
	}
	c.rooms[room.Code] = room
	return nil
}

func (c *MemCache) Delete(code string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.delete(code)
}

func (c *MemCache) delete(code string) {
	delete(c.rooms, code)
}

func (c *MemCache) StartMaintenanceJobs(stop <-chan struct{}) {
	c.cleanupJobOnce.Do(func() {
		ticker := time.NewTicker(cleanupInterval)
		go func() {
			for {
				select {
				case <-ticker.C:
					c.cleanupJob()
				case <-stop:
					ticker.Stop()
					return
				}
			}
		}()
	})

	c.shrinkJobOnce.Do(func() {
		ticker := time.NewTicker(downsizeCheckInterval)
		go func() {
			for {
				select {
				case <-ticker.C:
					c.shrinkIfRequired()
				case <-stop:
					ticker.Stop()
					return
				}
			}
		}()
	})
}

func (c *MemCache) cleanupJob() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for code, room := range c.rooms {
		if time.Since(room.CreatedTime) > cleanupInterval {

			c.delete(code)

		}
	}
}

func (c *MemCache) shrinkIfRequired() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.rooms) < downsizeThreshold {
		c.shrink()
	}
}

func (c *MemCache) shrink() {
	c.mu.Lock()
	defer c.mu.Unlock()

	newRooms := make(map[string]Room, len(c.rooms))

	for code, room := range c.rooms {
		newRooms[code] = room
	}

	c.rooms = newRooms
}
