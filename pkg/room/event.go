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
	EventTypeRoomReady   EventType = "room_ready"
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
	LastValidMoveCount int    `json:"lastValidMoveCount"`
	Error              string `json:"error"`
}

type EventResignPayload struct {
	Resigner engine.Color `json:"resigner"`
	Winner   engine.Color `json:"winner"`
}

type EventRoomReadyPayload struct {
	WhitePlayerName string `json:"whitePlayerName"`
	BlackPlayerName string `json:"blackPlayerName"`
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

func NewEventError(lastValidMoveCount int, err error) Event {
	return Event{
		EventType: EventTypeError,
		Payload: EventErrorPayload{
			LastValidMoveCount: lastValidMoveCount,
			Error:              err.Error(),
		},
	}
}

func NewEventResign(resigner engine.Color) Event {
	return Event{
		EventType: EventTypeResign,
		Payload: EventResignPayload{
			Resigner: resigner,
			Winner:   resigner.Opposite(),
		},
	}
}

func NewEventRoomReady(
	whitePlayerName string,
	blackPlayerName string,
) Event {
	return Event{
		EventType: EventTypeRoomReady,
		Payload: EventRoomReadyPayload{
			WhitePlayerName: whitePlayerName,
			BlackPlayerName: blackPlayerName,
		},
	}
}
