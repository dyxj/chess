package room

import (
	"encoding/json"

	"github.com/dyxj/chess/pkg/engine"
	"github.com/dyxj/chess/pkg/game"
)

type EventType string

const (
	EventTypeMessage     EventType = "message"
	EventTypeRoundResult EventType = "round"
	EventTypeError       EventType = "error"
	EventTypeResign      EventType = "resign"
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

type EventErrorPayload struct {
	Error string `json:"error"`
}

type EventResignPayload struct {
	Resigner engine.Color `json:"resigner"`
	Winner   engine.Color `json:"winner"`
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

func NewResignEvent(resigner engine.Color) Event {
	return Event{
		EventType: EventTypeResign,
		Payload: EventResignPayload{
			Resigner: resigner,
			Winner:   resigner.Opposite(),
		},
	}
}
