package room

import (
	"encoding/json"

	"github.com/dyxj/chess/internal/game"
)

type EventType string

const (
	EventTypeMessage     EventType = "message"
	EventTypeRoundResult EventType = "round"
	EventTypeError       EventType = "error"
)

type EventPartial struct {
	EventType EventType       `json:"type"`
	Payload   json.RawMessage `json:"payload"`
}

type Event struct {
	EventType EventType `json:"type"`
	Payload   any       `json:"payload"`
}
type EventMessagePayload struct {
	Message string `json:"message"`
}

type EventErrorCode string

const (
	EventErrorCodeInvalidMove    EventErrorCode = "invalid_move"
	EventErrorCodeNotActiveColor EventErrorCode = "not_active_color"
	EventErrorCodeIllegalMove    EventErrorCode = "illegal_move"
)

type EventErrorPayload struct {
	Error string `json:"error"`
}

func NewEventMessage(message string) Event {
	return Event{
		EventType: EventTypeMessage,
		Payload: EventMessagePayload{
			Message: message,
		},
	}
}

func NewEventRound(round game.RoundResult) Event {
	return Event{
		EventType: EventTypeRoundResult,
		Payload:   round,
	}
}

func NewEventError(err error) Event {
	return Event{
		EventType: EventTypeError,
		Payload: EventErrorPayload{
			Error: err.Error(),
		},
	}
}
