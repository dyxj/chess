package room

import (
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

type Room struct {
	ID          uuid.UUID  `json:"id"`
	Code        string     `json:"code"`
	Game        *game.Game `json:"game"`
	WhitePlayer Player     `json:"whitePlayer"`
	BlackPlayer Player     `json:"blackPlayer"`
}
