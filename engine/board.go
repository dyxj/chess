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

/*
NewBoard creates a new chess board with the initial pieces.

|110. 7|111. 7|112. 7|113. 7|114. 7|115. 7|116. 7|117. 7|118. 7|119. 7|
|100. 7|101. 7|102. 7|103. 7|104. 7|105. 7|106. 7|107. 7|108. 7|109. 7|
| 90. 7| 91.-4| 92.-2| 93.-3| 94.-5| 95.-6| 96.-3| 97.-2| 98.-4| 99. 7|
| 80. 7| 81.-1| 82.-1| 83.-1| 84.-1| 85.-1| 86.-1| 87.-1| 88.-1| 89. 7|
| 70. 7| 71. 0| 72. 0| 73. 0| 74. 0| 75. 0| 76. 0| 77. 0| 78. 0| 79. 7|
| 60. 7| 61. 0| 62. 0| 63. 0| 64. 0| 65. 0| 66. 0| 67. 0| 68. 0| 69. 7|
| 50. 7| 51. 0| 52. 0| 53. 0| 54. 0| 55. 0| 56. 0| 57. 0| 58. 0| 59. 7|
| 40. 7| 41. 0| 42. 0| 43. 0| 44. 0| 45. 0| 46. 0| 47. 0| 48. 0| 49. 7|
| 30. 7| 31. 1| 32. 1| 33. 1| 34. 1| 35. 1| 36. 1| 37. 1| 38. 1| 39. 7|
| 20. 7| 21. 4| 22. 2| 23. 3| 24. 5| 25. 6| 26. 3| 27. 2| 28. 4| 29. 7|
| 10. 7| 11. 7| 12. 7| 13. 7| 14. 7| 15. 7| 16. 7| 17. 7| 18. 7| 19. 7|
|  0. 7|  1. 7|  2. 7|  3. 7|  4. 7|  5. 7|  6. 7|  7. 7|  8. 7|  9. 7|
*/
func NewBoard() *Board {
	cells := [boardSize]int{}
	for i := 0; i < boardSize; i++ {
		cells[i] = calculateEmptyBoardValues(i)
	}
	b := &Board{
		cells: cells,
	}

	wp := GenerateStartPieces(White)
	err := b.LoadPieces(wp)
	if err != nil {
		panic(err)
	}

	bp := GenerateStartPieces(Black)
	err = b.LoadPieces(bp)
	if err != nil {
		panic(err)
	}

	return b
}

func NewEmptyBoard() *Board {
	cells := [boardSize]int{}
	for i := 0; i < boardSize; i++ {
		cells[i] = calculateEmptyBoardValues(i)
	}
	b := &Board{
		cells: cells,
	}
	return b
}

func (b *Board) LoadPieces(pp []*Piece) error {
	for _, p := range pp {
		err := b.setPiece(p)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Board) setPiece(p *Piece) error {
	if b.IsSentinel(p.position) {
		return ErrOutOfBoard
	}
	if !b.IsEmpty(p.position) {
		return ErrOccupied
	}
	b.cells[p.position] = boardSymbolPiece(p)
	return nil
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
		sb.WriteString("|")
		for y := 0; y < boardWidth; y++ {
			i := x*boardWidth + y

			sb.WriteString(fmt.Sprintf("%2d|", b.Value(i)))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func (b *Board) applyMove(m Move) {
	b.cells[m.From] = EmptyCell
	b.cells[m.To] = boardSymbolMove(&m)
}

func (b *Board) undoMove(move Move) {
	b.cells[move.From] = boardSymbolMove(&move)
	if move.Captured != 0 {
		b.cells[move.To] = int(move.Captured) * int(-move.Color)
	}
	b.cells[move.To] = EmptyCell
}

func (b *Board) undoMoves(moves []Move) {
	for i := len(moves) - 1; i >= 0; i-- {
		b.undoMove(moves[i])
	}
}

func boardSymbolPiece(p *Piece) int {
	return int(p.symbol) * int(p.color)
}

func boardSymbolMove(m *Move) int {
	return int(m.Symbol) * int(m.Color)
}

func calculateEmptyBoardValues(i int) int {
	if (i >= 0 && i <= 19) ||
		(i%10 == 0) ||
		(i%10 == 9) ||
		(i >= 100 && i <= 119) {
		return SentinelCell
	}

	return EmptyCell
}
