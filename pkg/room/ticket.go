package room

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/dyxj/chess/pkg/engine"
)

type PlayerTicket struct {
	RoomCode string
	Name     string
	Color    engine.Color
}

type TicketCache struct {
	cache sync.Map
}

func NewTicketCache() *TicketCache {
	return &TicketCache{}
}

func (c *TicketCache) GenerateTicket(
	roomCode string,
	name string,
	color engine.Color,
	duration time.Duration,
) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)

	c.cache.Store(token, PlayerTicket{RoomCode: roomCode, Name: name, Color: color})

	time.AfterFunc(duration, func() {
		c.cache.Delete(token)
	})

	return token
}

func (c *TicketCache) ConsumeTicket(token string) (PlayerTicket, bool) {
	val, ok := c.cache.LoadAndDelete(token)
	if !ok {
		return PlayerTicket{}, false
	}
	return val.(PlayerTicket), true
}
