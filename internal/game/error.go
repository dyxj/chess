package game

import "errors"

var ErrIllegalMove = errors.New("illegal move")
var ErrNotEligibleToForceDraw = errors.New("not eligible to force draw")
