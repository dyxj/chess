package room

import "errors"

var ErrCodeAlreadyExists = errors.New("code already exists")
var ErrRoomNotFound = errors.New("room not found")
var ErrRoomFull = errors.New("room is full")
var ErrColorOccupied = errors.New("color is occupied")
var ErrInvalidToken = errors.New("invalid token")

const (
	ErrCodeInvalidToken  = "invalid_token"
	ErrCodeRoomNotFound  = "room_not_found"
	ErrCodeRoomFull      = "room_full"
	ErrCodeColorOccupied = "color_occupied"
)

func terminalErrCode(err error) string {
	switch {
	case errors.Is(err, ErrInvalidToken):
		return ErrCodeInvalidToken
	case errors.Is(err, ErrRoomNotFound):
		return ErrCodeRoomNotFound
	case errors.Is(err, ErrRoomFull):
		return ErrCodeRoomFull
	case errors.Is(err, ErrColorOccupied):
		return ErrCodeColorOccupied
	default:
		return ""
	}
}
