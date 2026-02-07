package room

import "errors"

var ErrCodeAlreadyExists = errors.New("code already exists")
var ErrRoomNotFound = errors.New("room not found")
