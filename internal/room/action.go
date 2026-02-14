package room

import (
	"encoding/json"
	"errors"

	"github.com/dyxj/chess/internal/engine"
	"github.com/dyxj/chess/internal/game"
)

type ActionType string

const (
	ActionTypeMove ActionType = "move"
)

type ActionPartial struct {
	Type    ActionType      `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// ActionMovePayload
// coordinates are pointers as 0 is a valid value
type ActionMovePayload struct {
	Symbol engine.Symbol `json:"symbol"`
	From   *int          `json:"from"`
	To     *int          `json:"to"`
}

func (p *ActionMovePayload) Validate() error {
	if p.From == nil {
		return errors.New("from required")
	}

	if p.To == nil {
		return errors.New("to required")
	}

	if p.Symbol == 0 {
		return errors.New("symbol required")
	}

	return nil
}

func (p *ActionMovePayload) ToMove(color engine.Color) game.Move {
	return game.Move{
		Color:  color,
		Symbol: p.Symbol,
		From:   *p.From,
		To:     *p.To,
	}
}
