package engine_test

import (
	"testing"

	. "github.com/dyxj/chess/engine"
	"github.com/dyxj/chess/test/faker"
	"github.com/stretchr/testify/assert"
)

func TestRookPseudoLegalMoves(t *testing.T) {
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

				moves = append(moves, generateExpectedMoves(Rook, color, 54, 94, N)...)
				moves = append(moves, generateExpectedMoves(Rook, color, 54, 24, S)...)
				moves = append(moves, generateExpectedMoves(Rook, color, 54, 58, E)...)
				moves = append(moves, generateExpectedMoves(Rook, color, 54, 51, W)...)

				return moves
			},
		},
		{
			name: "blocked by same color pieces",
			otherPieces: func() []*Piece {
				return []*Piece{
					NewPiece(Rook, color, 74),
					NewPiece(Rook, color, 44),
					NewPiece(Rook, color, 57),
					NewPiece(Rook, color, 52),
				}
			},
			expect: func() []Move {
				var moves []Move

				moves = append(moves, generateExpectedMoves(Rook, color, 54, 64, N)...)
				moves = append(moves, generateExpectedMoves(Rook, color, 54, 54, S)...)
				moves = append(moves, generateExpectedMoves(Rook, color, 54, 56, E)...)
				moves = append(moves, generateExpectedMoves(Rook, color, 54, 53, W)...)

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
				}
			},
			expect: func() []Move {

				var moves []Move

				moves = append(moves, generateExpectedMoves(Rook, color, 54, 84, N, Rook)...)
				moves = append(moves, generateExpectedMoves(Rook, color, 54, 44, S, Rook)...)
				moves = append(moves, generateExpectedMoves(Rook, color, 54, 55, E, Rook)...)
				moves = append(moves, generateExpectedMoves(Rook, color, 54, 52, W, Rook)...)

				return moves
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			board := NewEmptyBoard()
			tPiece := NewPiece(Rook, color, 54)
			err := board.LoadPieces(
				append(tc.otherPieces(), tPiece),
			)
			assert.NoError(t, err)
			moves, err := GeneratePseudoLegalMoves(board, tPiece)
			assert.NoError(t, err)
			assert.Equal(t, tc.expect(), moves)
		})
	}
}
