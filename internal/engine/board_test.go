package engine

import (
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBoard(t *testing.T) {
	b := NewBoard()

	expectedCells := [120]int{
		7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
		7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
		7, 4, 2, 3, 5, 6, 3, 2, 4, 7,
		7, 1, 1, 1, 1, 1, 1, 1, 1, 7,
		7, 0, 0, 0, 0, 0, 0, 0, 0, 7,
		7, 0, 0, 0, 0, 0, 0, 0, 0, 7,
		7, 0, 0, 0, 0, 0, 0, 0, 0, 7,
		7, 0, 0, 0, 0, 0, 0, 0, 0, 7,
		7, -1, -1, -1, -1, -1, -1, -1, -1, 7,
		7, -4, -2, -3, -5, -6, -3, -2, -4, 7,
		7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
		7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
	}

	assert.Equal(t, 120, len(b.cells))
	assert.Equal(t, expectedCells, b.cells)
	assert.Equal(t, b.activeColor, White)
}

func TestNewEmptyBoard(t *testing.T) {
	b := NewEmptyBoard(Black)

	expectedCells := [120]int{
		7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
		7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
		7, 0, 0, 0, 0, 0, 0, 0, 0, 7,
		7, 0, 0, 0, 0, 0, 0, 0, 0, 7,
		7, 0, 0, 0, 0, 0, 0, 0, 0, 7,
		7, 0, 0, 0, 0, 0, 0, 0, 0, 7,
		7, 0, 0, 0, 0, 0, 0, 0, 0, 7,
		7, 0, 0, 0, 0, 0, 0, 0, 0, 7,
		7, 0, 0, 0, 0, 0, 0, 0, 0, 7,
		7, 0, 0, 0, 0, 0, 0, 0, 0, 7,
		7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
		7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
	}

	assert.Equal(t, 120, len(b.cells))
	assert.Equal(t, expectedCells, b.cells)
	assert.Equal(t, b.activeColor, Black)
}

func TestBoard_IsEmpty(t *testing.T) {
	board := genTestBoard()

	assert.True(t, board.IsEmpty(0))
	assert.False(t, board.IsEmpty(1))
	assert.False(t, board.IsEmpty(2))
	assert.False(t, board.IsEmpty(3))
}

func TestBoard_IsSentinel(t *testing.T) {
	board := genTestBoard()

	assert.False(t, board.IsSentinel(0))
	assert.True(t, board.IsSentinel(1))
	assert.False(t, board.IsSentinel(2))
	assert.False(t, board.IsEmpty(3))
}

func TestLoadPiecesError(t *testing.T) {
	tt := []struct {
		name     string
		pieces   func() []Piece
		expected error
	}{
		{
			name: "out of board(sentinel)",
			pieces: func() []Piece {
				var pieces []Piece
				pieces = append(pieces, NewPiece(Pawn, White, 99))
				return pieces
			},
			expected: ErrOutOfBoard,
		},
		{
			name: "out of board(index out of range)",
			pieces: func() []Piece {
				var pieces []Piece
				pieces = append(pieces, NewPiece(Pawn, White, 120))
				return pieces
			},
			expected: ErrOutOfBoard,
		},
		{
			name: "occupied",
			pieces: func() []Piece {
				var pieces []Piece
				pieces = append(pieces, NewPiece(Rook, White, 74))
				pieces = append(pieces, NewPiece(Pawn, White, 74))
				return pieces
			},
			expected: ErrOccupied,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			board := NewEmptyBoard()
			err := board.LoadPieces(tc.pieces())
			assert.Equal(t, tc.expected, err)
		})
	}
}

func genTestBoard() *Board {
	return &Board{
		cells: [120]int{
			0,
			7,
			rand.IntN(5) + 1,
			(rand.IntN(5) + 1) * -1,
		},
	}
}
