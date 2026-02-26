package room

import (
	"encoding/json"
	"errors"
	"slices"

	"github.com/dyxj/chess/pkg/engine"
	"github.com/dyxj/chess/pkg/game"
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
	Symbol    engine.Symbol `json:"symbol"`
	From      *int          `json:"from"`
	To        *int          `json:"to"`
	Promotion engine.Symbol `json:"promotion"`
}

func (p *ActionMovePayload) Validate() error {
	if p.From == nil {
		return errors.New("from required")
	}
	if *p.From < 0 || *p.From > 63 {
		return errors.New("from must be between 0 and 63")
	}

	if p.To == nil {
		return errors.New("to required")
	}
	if *p.To < 0 || *p.To > 63 {
		return errors.New("to must be between 0 and 63")
	}

	if !slices.Contains(engine.Symbols, p.Symbol) {
		return errors.New("invalid symbol")
	}

	if p.Promotion != 0 && !slices.Contains(engine.PromotionSymbols, p.Promotion) {
		return errors.New("invalid promotion symbol")
	}

	return nil
}

func (p *ActionMovePayload) ToMove(color engine.Color) game.Move {
	return game.Move{
		Color:     color,
		Symbol:    p.Symbol,
		From:      *p.From,
		To:        *p.To,
		Promotion: p.Promotion,
	}
}
