package game

import "errors"

var ErrIllegalMove = errors.New("illegal move")
var ErrInvalidMove = errors.New("invalid move")
var ErrNotEligibleToForceDraw = errors.New("not eligible to force draw")
