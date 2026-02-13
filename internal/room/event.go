package room

import "encoding/json"

type EventType string

const (
	EventTypeMessage EventType = "message"
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

func NewEventMessage(message string) Event {
	return Event{
		EventType: EventTypeMessage,
		Payload: EventMessagePayload{
			Message: message,
		},
	}
}

type ActionType string

type Action struct {
	Action  ActionType      `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

//type Event struct {
//	Status    Status     `json:"status"`
//	Message   string     `json:"message"`
//	GameState game.State `json:"gameState"`
//	Move      game.Move  `json:"move"`
//}

//type ActionType int
//
//const (
//	ActionTypeMove ActionType = iota
//	ActionTypeDraw
//	ActionTypeResign
//)
//
//type Action struct {
//	Type ActionType `json:"type"`
//	From *int       `json:"from"`
//	To   *int       `json:"to"`
//}
