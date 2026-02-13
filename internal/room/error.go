package room

import "errors"

var ErrCodeAlreadyExists = errors.New("code already exists")
var ErrRoomNotFound = errors.New("room not found")
var ErrRoomFull = errors.New("room is full")
var ErrColorOccupied = errors.New("color is occupied")
