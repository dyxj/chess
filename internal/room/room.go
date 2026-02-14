package room

import (
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
	whitePlayer   *Player
	blackPlayer   *Player
	CreatedTime   time.Time
	ticketsIssued int
	readyChan     chan struct{}
	readyOnce     sync.Once
	gameOverChan  chan struct{}
	gameOverOnce  sync.Once
}

func NewEmptyRoom() *Room {
	return &Room{
		ID:            uuid.New(),
		Code:          generateCode(),
		Game:          game.NewGame(engine.NewBoard()),
		status:        StatusWaiting,
		CreatedTime:   time.Now(),
		ticketsIssued: 0,
		readyChan:     make(chan struct{}),
		gameOverChan:  make(chan struct{}),
	}
}

func (r *Room) Player(color engine.Color) *Player {
	if color == engine.White {
		return r.whitePlayer
	}
	return r.blackPlayer
}

func (r *Room) SetPlayer(color engine.Color, p *Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if color == engine.White {
		if r.whitePlayer != nil {
			return ErrColorOccupied
		}
		r.whitePlayer = p
	} else {
		if r.blackPlayer != nil {
			return ErrColorOccupied
		}
		r.blackPlayer = p
	}
	return nil
}

func (r *Room) RemovePlayer(color engine.Color) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if color == engine.White {
		r.whitePlayer = nil
	} else {
		r.blackPlayer = nil
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

func (r *Room) HasBothPlayers() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.whitePlayer != nil && r.blackPlayer != nil {
		return true
	}
	return false
}

func (r *Room) signalReady() {
	r.readyOnce.Do(func() {
		close(r.readyChan)
	})
}

func (r *Room) signalGameOver() {
	r.gameOverOnce.Do(func() {
		close(r.gameOverChan)
	})
}
