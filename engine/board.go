package engine

import "fmt"

// Board dimensions
const boardWidth = 10  // column
const boardHeight = 12 // row

type Direction int

const (
	N  Direction = 10
	S  Direction = -10
	E  Direction = 1
	W  Direction = -1
	NE Direction = N + E // 11
	NW Direction = N + W // 9
	SE Direction = S + E // -9
	SW Direction = S + W // -11
)

// Board representation
const (
	EmptyCell    = 0
	SentinelCell = 7
)

type Board struct {
	cell [width * height]int
}

func NewBoard() *Board {
	// initialize new board
	return &Board{}
}

func (b *Board) IsEmpty(i int) bool {
	return b.cell[i] == EmptyCell
}
func (b *Board) IsSentinel(i int) bool {
	return b.cell[i] == SentinelCell
}

func (b *Board) Color(i int) Color {
	cv := b.cell[i]
	if cv == 0 || cv == 7 {
		// risky silent failure
		// options
		// 1. could return ok check
		// 2. log error
		// 3. panic
		// Choosing to ignore relying on right implementation
		return 0
	}
	if cv > 0 {
		return White
	}
	return Black
}

func (b *Board) Symbol(i int) Symbol {
	cv := b.cell[i]
	if cv == 0 || cv == 7 {
		// risky silent failure
		// options
		// 1. could return ok check
		// 2. log error
		// 3. panic
		// Choosing to ignore relying on right implementation
		return 0
	}
	if cv > 0 {
		return Symbol(cv)
	}
	return Symbol(-cv)
}
