package room

import (
	"encoding/json"

	"github.com/dyxj/chess/internal/engine"
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
	Color  engine.Color  `json:"color"`
	Symbol engine.Symbol `json:"symbol"`
	From   *int          `json:"from"`
	To     *int          `json:"to"`
}
