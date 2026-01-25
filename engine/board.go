package engine

import (
	"fmt"
	"strings"
)

// Board dimensions
const boardWidth = 10  // column
const boardHeight = 12 // row
const boardSize = boardWidth * boardHeight

// Board representation
const (
	EmptyCell    = 0
	SentinelCell = 7
)

type Board struct {
	cells [boardSize]int
}

func NewBoard() *Board {
	cells := [boardSize]int{}
	for i := 0; i < boardSize; i++ {
		cells[i] = resolveInitialBoardCellValue(i)
	}
	return &Board{
		cells: cells,
	}
}

func (b *Board) IsEmpty(i int) bool {
	return b.cells[i] == EmptyCell
}
func (b *Board) IsSentinel(i int) bool {
	return b.cells[i] == SentinelCell
}

func (b *Board) Color(i int) Color {
	cv := b.cells[i]
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
	cv := b.cells[i]
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

func (b *Board) Value(i int) int {
	return b.cells[i]
}

func (b *Board) GridString() string {
	sb := strings.Builder{}
	for x := boardHeight - 1; x >= 0; x-- {
		for y := 0; y < boardWidth; y++ {
			i := x*boardWidth + y
			sb.WriteString(fmt.Sprintf("(%2d) ", b.Value(i)))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func resolveInitialBoardCellValue(i int) int {
	if (i >= 0 && i <= 19) ||
		(i%10 == 0) ||
		(i%10 == 9) ||
		(i >= 100 && i <= 119) {
		return SentinelCell
	}

	if i >= 31 && i <= 38 {
		return int(Pawn) * int(White)
	}

	if i >= 81 && i <= 88 {
		return int(Pawn) * int(Black)
	}

	var powerPiece Symbol
	switch i {
	case 21, 28, 91, 98:
		powerPiece = Rook
	case 22, 27, 92, 97:
		powerPiece = Knight
	case 23, 26, 93, 96:
		powerPiece = Bishop
	case 24, 94:
		powerPiece = Queen
	case 25, 95:
		powerPiece = King
	}

	if powerPiece == 0 {
		return EmptyCell
	}

	if i < 30 {
		return int(powerPiece) * int(White)
	}
	return int(powerPiece) * int(Black)
}

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
