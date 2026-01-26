package engine_test

import (
	"testing"

	. "github.com/dyxj/chess/engine"
	"github.com/dyxj/chess/test/faker"
	"github.com/stretchr/testify/assert"
)

func TestKingBasicMoves(t *testing.T) {
	color := faker.Color()
	xColor := color.Opposite()

	tt := []struct {
		name        string
		otherPieces func() []*Piece
		expect      func() []Move
	}{
		{
			name: "all directions",
			otherPieces: func() []*Piece {
				return []*Piece{}
			},
			expect: func() []Move {
				var moves []Move

				// King moves only one square in each direction
				moves = append(moves, generateExpectedMoves(King, color, 54, 64, N)...)
				moves = append(moves, generateExpectedMoves(King, color, 54, 44, S)...)
				moves = append(moves, generateExpectedMoves(King, color, 54, 55, E)...)
				moves = append(moves, generateExpectedMoves(King, color, 54, 53, W)...)
				moves = append(moves, generateExpectedMoves(King, color, 54, 65, NE)...)
				moves = append(moves, generateExpectedMoves(King, color, 54, 63, NW)...)
				moves = append(moves, generateExpectedMoves(King, color, 54, 45, SE)...)
				moves = append(moves, generateExpectedMoves(King, color, 54, 43, SW)...)

				return moves
			},
		},
		{
			name: "blocked by same color pieces",
			otherPieces: func() []*Piece {
				return []*Piece{
					NewPiece(Pawn, color, 64),
					NewPiece(Pawn, color, 44),
					NewPiece(Pawn, color, 55),
					NewPiece(Pawn, color, 53),
					NewPiece(Pawn, color, 65),
					NewPiece(Pawn, color, 63),
					NewPiece(Pawn, color, 45),
					NewPiece(Pawn, color, 43),
				}
			},
			expect: func() []Move {
				// King cannot move to any square, all blocked
				var moves []Move
				return moves
			},
		},
		{
			name: "capture",
			otherPieces: func() []*Piece {
				return []*Piece{
					NewPiece(Pawn, xColor, 64),
					NewPiece(Pawn, xColor, 55),
					NewPiece(Pawn, xColor, 65),
					NewPiece(Pawn, xColor, 43),
				}
			},
			expect: func() []Move {
				var moves []Move

				moves = append(moves, generateExpectedMoves(King, color, 54, 64, N, Pawn)...)
				moves = append(moves, generateExpectedMoves(King, color, 54, 44, S)...)
				moves = append(moves, generateExpectedMoves(King, color, 54, 55, E, Pawn)...)
				moves = append(moves, generateExpectedMoves(King, color, 54, 53, W)...)
				moves = append(moves, generateExpectedMoves(King, color, 54, 65, NE, Pawn)...)
				moves = append(moves, generateExpectedMoves(King, color, 54, 63, NW)...)
				moves = append(moves, generateExpectedMoves(King, color, 54, 45, SE)...)
				moves = append(moves, generateExpectedMoves(King, color, 54, 43, SW, Pawn)...)

				return moves
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			board := NewEmptyBoard()
			tPiece := NewPiece(King, color, 54)
			err := board.LoadPieces(
				append(tc.otherPieces(), tPiece),
			)
			assert.NoError(t, err)
			moves := GenerateBasicMoves(board, tPiece)
			assert.Equal(t, tc.expect(), moves)
		})
	}
}

func TestKingEndOfBoard(t *testing.T) {
	color := faker.Color()
	tt := []struct {
		name                  string
		kingPosition          int
		expectedNumberOfMoves int
	}{
		{
			"top left corner",
			91,
			3,
		},
		{
			"top right corner",
			98,
			3,
		},
		{
			"top center",
			95,
			5,
		},
		{
			"bottom left corner",
			21,
			3,
		},
		{
			"bottom right corner",
			28,
			3,
		},
		{
			"bottom center",
			25,
			5,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			board := NewEmptyBoard()
			tPiece := NewPiece(King, color, tc.kingPosition)
			err := board.LoadPieces(
				append([]*Piece{}, tPiece),
			)
			assert.NoError(t, err)
			moves := GenerateBasicMoves(board, tPiece)
			assert.Equal(t, tc.expectedNumberOfMoves, len(moves))
		})
	}
}
