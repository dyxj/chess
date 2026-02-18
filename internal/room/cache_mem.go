package room

import (
	"time"

	"github.com/dyxj/chess/pkg/store"
)

const maximumRoomDuration = 6 * time.Hour

type MemCache struct {
	cache *store.MemCache
}

func NewMemCache(
	cache *store.MemCache,
) *MemCache {
	return &MemCache{
		cache: cache,
	}
}

func (c *MemCache) Add(room *Room) error {
	err := c.cache.Add(room.Code, room, time.Now().Add(maximumRoomDuration))
	if err != nil {
		return ErrCodeAlreadyExists
	}
	return nil
}

func (c *MemCache) Find(code string) (*Room, bool) {
	item, ok := c.cache.Find(code)
	if !ok {
		return nil, false
	}
	room, isRoom := item.(*Room)
	if !isRoom {
		return nil, false
	}
	return room, true
}
