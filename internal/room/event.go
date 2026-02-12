package room

import "github.com/dyxj/chess/internal/game"

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
