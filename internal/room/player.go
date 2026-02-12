package room

import "github.com/google/uuid"

type Player struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func NewPlayer() Player {
	return Player{
		ID: uuid.New(),
	}
}
