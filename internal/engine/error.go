package engine

import "errors"

var ErrOutOfBoard = errors.New("position out of board")
var ErrOccupied = errors.New("position is occupied")
var ErrPieceNotFound = errors.New("piece not found on the board")
var ErrNotActiveColor = errors.New("not active color")
