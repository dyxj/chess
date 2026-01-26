package engine_test

import (
	"testing"

	. "github.com/dyxj/chess/engine"
	"github.com/dyxj/chess/test/faker"
	"github.com/stretchr/testify/assert"
)

func TestQueenBasicMoves(t *testing.T) {
	color := faker.Color()
	xColor := color.Opposite()

	tt := []struct {
		name        string
		otherPieces func() []*Piece
		expect      func() []Move
	}{
		{
			name: "end of board",
			otherPieces: func() []*Piece {
				return []*Piece{}
			},
			expect: func() []Move {
				var moves []Move

				moves = append(moves, generateExpectedMoves(Queen, color, 54, 94, N)...)
				moves = append(moves, generateExpectedMoves(Queen, color, 54, 24, S)...)
				moves = append(moves, generateExpectedMoves(Queen, color, 54, 58, E)...)
				moves = append(moves, generateExpectedMoves(Queen, color, 54, 51, W)...)
				moves = append(moves, generateExpectedMoves(Queen, color, 54, 98, NE)...)
				moves = append(moves, generateExpectedMoves(Queen, color, 54, 81, NW)...)
				moves = append(moves, generateExpectedMoves(Queen, color, 54, 27, SE)...)
				moves = append(moves, generateExpectedMoves(Queen, color, 54, 21, SW)...)

				return moves
			},
		},
		{
			name: "blocked by same color pieces",
			otherPieces: func() []*Piece {
				return []*Piece{
					NewPiece(Pawn, color, 74),
					NewPiece(Pawn, color, 44),
					NewPiece(Pawn, color, 57),
					NewPiece(Pawn, color, 52),
					NewPiece(Pawn, color, 76),
					NewPiece(Pawn, color, 72),
					NewPiece(Pawn, color, 36),
					NewPiece(Pawn, color, 32),
				}
			},
			expect: func() []Move {
				var moves []Move

				moves = append(moves, generateExpectedMoves(Queen, color, 54, 64, N)...)
				moves = append(moves, generateExpectedMoves(Queen, color, 54, 54, S)...)
				moves = append(moves, generateExpectedMoves(Queen, color, 54, 56, E)...)
				moves = append(moves, generateExpectedMoves(Queen, color, 54, 53, W)...)
				moves = append(moves, generateExpectedMoves(Queen, color, 54, 65, NE)...)
				moves = append(moves, generateExpectedMoves(Queen, color, 54, 63, NW)...)
				moves = append(moves, generateExpectedMoves(Queen, color, 54, 45, SE)...)
				moves = append(moves, generateExpectedMoves(Queen, color, 54, 43, SW)...)

				return moves
			},
		},
		{
			name: "capture",
			otherPieces: func() []*Piece {
				return []*Piece{
					NewPiece(Rook, xColor, 84),
					NewPiece(Rook, xColor, 44),
					NewPiece(Rook, xColor, 55),
					NewPiece(Rook, xColor, 52),
					NewPiece(Bishop, xColor, 76),
					NewPiece(Bishop, xColor, 72),
					NewPiece(Bishop, xColor, 36),
					NewPiece(Bishop, xColor, 32),
				}
			},
			expect: func() []Move {
				var moves []Move

				moves = append(moves, generateExpectedMoves(Queen, color, 54, 84, N, Rook)...)
				moves = append(moves, generateExpectedMoves(Queen, color, 54, 44, S, Rook)...)
				moves = append(moves, generateExpectedMoves(Queen, color, 54, 55, E, Rook)...)
				moves = append(moves, generateExpectedMoves(Queen, color, 54, 52, W, Rook)...)
				moves = append(moves, generateExpectedMoves(Queen, color, 54, 76, NE, Bishop)...)
				moves = append(moves, generateExpectedMoves(Queen, color, 54, 72, NW, Bishop)...)
				moves = append(moves, generateExpectedMoves(Queen, color, 54, 36, SE, Bishop)...)
				moves = append(moves, generateExpectedMoves(Queen, color, 54, 32, SW, Bishop)...)

				return moves
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			board := NewEmptyBoard()
			tPiece := NewPiece(Queen, color, 54)
			err := board.LoadPieces(
				append(tc.otherPieces(), tPiece),
			)
			assert.NoError(t, err)
			moves := GenerateBasicMoves(board, tPiece)
			assert.Equal(t, tc.expect(), moves)
		})
	}
}
