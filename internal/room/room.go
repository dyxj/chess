package room

import (
	"slices"
	"sync"
	"time"

	"github.com/dyxj/chess/internal/engine"
	"github.com/dyxj/chess/internal/game"
	"github.com/google/uuid"
)

type Type int

const (
	TypePublic Type = iota
	TypePrivate
)

var Types = []Type{TypePublic, TypePrivate}

func (rt Type) String() string {
	switch rt {
	case TypePublic:
		return "public"
	case TypePrivate:
		return "private"
	default:
		return "unknown"
	}
}

type Status int

const (
	StatusWaiting Status = iota
	StatusInProgress
	StatusCompleted
)

func (s Status) String() string {
	switch s {
	case StatusWaiting:
		return "waiting"
	case StatusInProgress:
		return "in_progress"
	case StatusCompleted:
		return "completed"
	default:
		return "unknown"
	}
}

type Room struct {
	mu            sync.RWMutex
	ID            uuid.UUID
	Code          string
	status        Status
	Game          *game.Game
	Players       []Player
	CreatedTime   time.Time
	ticketsIssued int
}

func NewEmptyRoom() *Room {
	return &Room{
		ID:            uuid.New(),
		Code:          generateCode(),
		Game:          game.NewGame(engine.NewBoard()),
		status:        StatusWaiting,
		CreatedTime:   time.Now(),
		Players:       make([]Player, 0, 2),
		ticketsIssued: 0,
	}
}

func (r *Room) AddPlayer(p Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.Players) >= 2 {
		return ErrRoomFull
	}
	r.Players = append(r.Players, p)
	return nil
}

func (r *Room) RemovePlayer(p Player) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, player := range r.Players {
		if player.ID == p.ID {
			r.Players = slices.Delete(r.Players, i, i+1)
			return
		}
	}
}

func (r *Room) IncrementTicket() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ticketsIssued >= 2 {
		return ErrRoomFull
	}
	r.ticketsIssued++
	return nil
}

func (r *Room) DecrementTicket() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ticketsIssued > 0 {
		r.ticketsIssued--
	}
}

func (r *Room) Status() Status {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.status
}

func (r *Room) SetStatus(s Status) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.status = s
}
