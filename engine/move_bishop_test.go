package engine_test

import (
	"testing"

	. "github.com/dyxj/chess/engine"
	"github.com/dyxj/chess/test/faker"
	"github.com/stretchr/testify/assert"
)

func TestBishopPseudoLegalMoves(t *testing.T) {
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

				moves = append(moves, generateExpectedMoves(Bishop, color, 54, 98, NE)...)
				moves = append(moves, generateExpectedMoves(Bishop, color, 54, 81, NW)...)
				moves = append(moves, generateExpectedMoves(Bishop, color, 54, 27, SE)...)
				moves = append(moves, generateExpectedMoves(Bishop, color, 54, 21, SW)...)

				return moves
			},
		},
		{
			name: "blocked by same color pieces",
			otherPieces: func() []*Piece {
				return []*Piece{
					NewPiece(Bishop, color, 76),
					NewPiece(Bishop, color, 72),
					NewPiece(Bishop, color, 36),
					NewPiece(Bishop, color, 32),
				}
			},
			expect: func() []Move {
				var moves []Move

				moves = append(moves, generateExpectedMoves(Bishop, color, 54, 65, NE)...)
				moves = append(moves, generateExpectedMoves(Bishop, color, 54, 63, NW)...)
				moves = append(moves, generateExpectedMoves(Bishop, color, 54, 45, SE)...)
				moves = append(moves, generateExpectedMoves(Bishop, color, 54, 43, SW)...)

				return moves
			},
		},
		{
			name: "capture",
			otherPieces: func() []*Piece {
				return []*Piece{
					NewPiece(Bishop, xColor, 76),
					NewPiece(Bishop, xColor, 72),
					NewPiece(Bishop, xColor, 36),
					NewPiece(Bishop, xColor, 32),
				}
			},
			expect: func() []Move {

				var moves []Move

				moves = append(moves, generateExpectedMoves(Bishop, color, 54, 76, NE, Bishop)...)
				moves = append(moves, generateExpectedMoves(Bishop, color, 54, 72, NW, Bishop)...)
				moves = append(moves, generateExpectedMoves(Bishop, color, 54, 36, SE, Bishop)...)
				moves = append(moves, generateExpectedMoves(Bishop, color, 54, 32, SW, Bishop)...)

				return moves
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			board := NewEmptyBoard()
			tPiece := NewPiece(Bishop, color, 54)
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
