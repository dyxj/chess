package engine

import (
	"testing"

	"github.com/dyxj/chess/pkg/randx"
	"github.com/stretchr/testify/assert"
)

// Test filtering of moves
func TestGeneratePieceLegalMoves(t *testing.T) {
	color := randx.FromSlice(Colors)
	xColor := color.Opposite()

	tt := []struct {
		name        string
		otherPieces func() []Piece
		expect      func() []Move
	}{
		{
			name: "all moves without king on board",
			otherPieces: func() []Piece {
				var pieces []Piece

				pieces = append(pieces, NewPiece(Rook, xColor, 24, true))

				return pieces
			},
			expect: func() []Move {
				var moves []Move

				moves = append(moves, generateExpectedMoves(Rook, color, 54, 94, N)...)
				moves = append(moves, generateExpectedMoves(Rook, color, 54, 24, S, Rook)...)
				moves = append(moves, generateExpectedMoves(Rook, color, 54, 58, E)...)
				moves = append(moves, generateExpectedMoves(Rook, color, 54, 51, W)...)

				return moves
			},
		},
		{
			name: "all moves",
			otherPieces: func() []Piece {
				var pieces []Piece

				pieces = append(pieces, NewPiece(King, color, 65, true))
				pieces = append(pieces, NewPiece(Rook, xColor, 24, true))

				return pieces
			},
			expect: func() []Move {
				var moves []Move

				moves = append(moves, generateExpectedMoves(Rook, color, 54, 94, N)...)
				moves = append(moves, generateExpectedMoves(Rook, color, 54, 24, S, Rook)...)
				moves = append(moves, generateExpectedMoves(Rook, color, 54, 58, E)...)
				moves = append(moves, generateExpectedMoves(Rook, color, 54, 51, W)...)

				return moves
			},
		},
		{
			name: "horizontal moves not possible due to check",
			otherPieces: func() []Piece {
				var pieces []Piece

				pieces = append(pieces, NewPiece(King, color, 74, true))
				pieces = append(pieces, NewPiece(Rook, xColor, 24, true))

				return pieces
			},
			expect: func() []Move {
				var moves []Move

				moves = append(moves, generateExpectedMoves(Rook, color, 54, 64, N)...)
				moves = append(moves, generateExpectedMoves(Rook, color, 54, 24, S, Rook)...)

				return moves
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			board := NewEmptyBoard()
			tPiece := NewPiece(Rook, color, 54, true)
			err := board.LoadPieces(
				append(tc.otherPieces(), tPiece),
			)
			assert.NoError(t, err)
			moves, err := board.GeneratePieceLegalMoves(tPiece)
			assert.NoError(t, err)
			assert.Equal(t, tc.expect(), moves)
		})
	}
}

func TestGeneratePieceLegalMovesErrors(t *testing.T) {
	s := Pawn
	c := randx.FromSlice(Colors)
	board := NewEmptyBoard()

	err := board.loadPiece(NewPiece(s, c, 54, true))
	assert.NoError(t, err)

	_, err = board.GeneratePieceLegalMoves(NewPiece(Rook, c, 54))
	assert.ErrorIs(t, err, ErrPieceNotFound)
	_, err = board.GeneratePieceLegalMoves(NewPiece(Pawn, c.Opposite(), 54))
	assert.ErrorIs(t, err, ErrPieceNotFound)
	_, err = board.GeneratePieceLegalMoves(NewPiece(s, c, 55))
	assert.ErrorIs(t, err, ErrPieceNotFound)
}

func TestGenerateLegalMoves(t *testing.T) {
	color := randx.FromSlice(Colors)
	xColor := color.Opposite()

	tt := []struct {
		name        string
		otherPieces func() []Piece
		expect      func() []Move
	}{
		{
			name: "all moves except king self check",
			otherPieces: func() []Piece {
				var pieces []Piece

				pieces = append(pieces, NewPiece(King, color, 55, true))
				pieces = append(pieces, NewPiece(Rook, xColor, 24, true))

				return pieces
			},
			expect: func() []Move {
				var moves []Move

				moves = append(moves, generateExpectedMoves(King, color, 55, 65, N)...)
				moves = append(moves, generateExpectedMoves(King, color, 55, 45, S)...)
				moves = append(moves, generateExpectedMoves(King, color, 55, 56, E)...)
				// W excluded due to block 54
				moves = append(moves, generateExpectedMoves(King, color, 55, 66, NE)...)
				moves = append(moves, generateExpectedMoves(King, color, 55, 46, SE)...)
				moves = append(moves, generateExpectedMoves(King, color, 55, 64, NW)...)
				// SW excluded due to check 44

				moves = append(moves, generateExpectedMoves(Rook, color, 54, 94, N)...)
				moves = append(moves, generateExpectedMoves(Rook, color, 54, 24, S, Rook)...)
				// E excluded due to block 55
				moves = append(moves, generateExpectedMoves(Rook, color, 54, 51, W)...)

				return moves
			},
		},
		{
			name: "horizontal rook moves not possible due to check",
			otherPieces: func() []Piece {
				var pieces []Piece

				pieces = append(pieces, NewPiece(King, color, 74, true))
				pieces = append(pieces, NewPiece(Rook, xColor, 24, true))

				return pieces
			},
			expect: func() []Move {
				var moves []Move

				moves = append(moves, generateExpectedMoves(King, color, 74, 84, N)...)
				moves = append(moves, generateExpectedMoves(King, color, 74, 64, S)...)
				moves = append(moves, generateExpectedMoves(King, color, 74, 75, E)...)
				moves = append(moves, generateExpectedMoves(King, color, 74, 73, W)...)
				moves = append(moves, generateExpectedMoves(King, color, 74, 85, NE)...)
				moves = append(moves, generateExpectedMoves(King, color, 74, 83, NW)...)
				moves = append(moves, generateExpectedMoves(King, color, 74, 65, SE)...)
				moves = append(moves, generateExpectedMoves(King, color, 74, 63, SW)...)

				moves = append(moves, generateExpectedMoves(Rook, color, 54, 64, N)...)
				moves = append(moves, generateExpectedMoves(Rook, color, 54, 24, S, Rook)...)

				return moves
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			board := NewEmptyBoard()
			tPiece := NewPiece(Rook, color, 54, true)
			err := board.LoadPieces(
				append(tc.otherPieces(), tPiece),
			)
			assert.NoError(t, err)
			moves := board.GenerateLegalMoves(color)

			expect := tc.expect()
			sortMoves(expect)
			sortMoves(moves)

			assert.Equal(t, expect, moves)
		})
	}
}

func TestGenerateLegalMovesErrors(t *testing.T) {
	s := Pawn
	c := randx.FromSlice(Colors)
	board := NewEmptyBoard()

	err := board.loadPiece(NewPiece(s, c, 54, true))
	assert.NoError(t, err)

	// programmer error, cells and pieces out of sync
	board.cells[54] = EmptyCell

	assert.PanicsWithError(t, ErrPieceNotFound.Error(), func() {
		_ = board.GenerateLegalMoves(c)
	})
}
