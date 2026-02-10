package room

import (
	"time"

	"github.com/dyxj/chess/internal/game"
	"github.com/dyxj/chess/pkg/websocketx"
	"github.com/google/uuid"
	"go.uber.org/zap"
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

type color string

const (
	white color = "white"
	black color = "black"
)

func (c color) String() string {
	return string(c)
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
		return "in progress"
	case StatusCompleted:
		return "completed"
	default:
		return "unknown"
	}
}

type Room struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Status      Status    `json:"status"`
	GameID      uuid.UUID `json:"gameId"`
	WhitePlayer Player    `json:"whitePlayer"`
	BlackPlayer Player    `json:"blackPlayer"`
	CreatedTime time.Time `json:"createdTime"`
}

func (r *Room) setPlayer(color color, player Player) {
	if color == white {
		r.WhitePlayer = player
	} else {
		r.BlackPlayer = player
	}
}

func (r *Room) player(color color) Player {
	if color == white {
		return r.WhitePlayer
	}
	return r.BlackPlayer
}

func (r *Room) connectionKey(color color) string {
	return r.ID.String() + ":" + color.String()
}

func NewEmptyRoom() *Room {
	return &Room{
		ID:          uuid.New(),
		Code:        generateCode(),
		Status:      StatusInProgress,
		CreatedTime: time.Now(),
	}
}

type Event struct {
	Status    Status     `json:"status"`
	Message   string     `json:"message"`
	GameState game.State `json:"gameState"`
	Move      game.Move  `json:"move"`
}

type ActionType int

const (
	ActionTypeMove ActionType = iota
	ActionTypeDraw
	ActionTypeResign
)

type Action struct {
	Type ActionType `json:"type"`
	From *int       `json:"from"`
	To   *int       `json:"to"`
}

func NewWebSocketManager(logger *zap.Logger) *websocketx.Manager {
	return websocketx.NewManager(logger)
}
