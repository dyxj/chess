package engine

import (
	"testing"

	"github.com/dyxj/chess/pkg/randx"
	"github.com/stretchr/testify/assert"
)

func TestKnightPseudoLegalMoves(t *testing.T) {
	color := randx.FromSlice(Colors)
	xColor := color.Opposite()

	NNE := N + N + E // 21
	NNW := N + N + W // 19
	SSE := S + S + E // -19
	SSW := S + S + W // -21
	EEN := E + E + N // 12
	EES := E + E + S // -8
	WWN := W + W + N // 8
	WWS := W + W + S // -12

	tt := []struct {
		name        string
		otherPieces func() []Piece
		expect      func() []Move
	}{
		{
			name: "all directions",
			otherPieces: func() []Piece {
				return []Piece{}
			},
			expect: func() []Move {
				var moves []Move

				moves = append(moves, generateExpectedMoves(Knight, color, 54, 75, NNE)...)
				moves = append(moves, generateExpectedMoves(Knight, color, 54, 73, NNW)...)
				moves = append(moves, generateExpectedMoves(Knight, color, 54, 35, SSE)...)
				moves = append(moves, generateExpectedMoves(Knight, color, 54, 33, SSW)...)
				moves = append(moves, generateExpectedMoves(Knight, color, 54, 66, EEN)...)
				moves = append(moves, generateExpectedMoves(Knight, color, 54, 46, EES)...)
				moves = append(moves, generateExpectedMoves(Knight, color, 54, 62, WWN)...)
				moves = append(moves, generateExpectedMoves(Knight, color, 54, 42, WWS)...)

				return moves
			},
		},
		{
			name: "blocked by same color pieces at final position",
			otherPieces: func() []Piece {
				return []Piece{
					NewPiece(Pawn, color, 75),
					NewPiece(Pawn, color, 73),
					NewPiece(Pawn, color, 35),
					NewPiece(Pawn, color, 33),
					NewPiece(Pawn, color, 66),
					NewPiece(Pawn, color, 46),
					NewPiece(Pawn, color, 62),
					NewPiece(Pawn, color, 42),
				}
			},
			expect: func() []Move {
				return []Move{}
			},
		},
		{
			name: "capture",
			otherPieces: func() []Piece {
				return []Piece{
					NewPiece(Pawn, xColor, 75),
					NewPiece(Pawn, xColor, 73),
					NewPiece(Pawn, xColor, 35),
					NewPiece(Pawn, xColor, 33),
					NewPiece(Pawn, xColor, 66),
					NewPiece(Pawn, xColor, 46),
					NewPiece(Pawn, xColor, 62),
					NewPiece(Pawn, xColor, 42),
				}
			},
			expect: func() []Move {
				var moves []Move

				moves = append(moves, generateExpectedMoves(Knight, color, 54, 75, NNE, Pawn)...)
				moves = append(moves, generateExpectedMoves(Knight, color, 54, 73, NNW, Pawn)...)
				moves = append(moves, generateExpectedMoves(Knight, color, 54, 35, SSE, Pawn)...)
				moves = append(moves, generateExpectedMoves(Knight, color, 54, 33, SSW, Pawn)...)
				moves = append(moves, generateExpectedMoves(Knight, color, 54, 66, EEN, Pawn)...)
				moves = append(moves, generateExpectedMoves(Knight, color, 54, 46, EES, Pawn)...)
				moves = append(moves, generateExpectedMoves(Knight, color, 54, 62, WWN, Pawn)...)
				moves = append(moves, generateExpectedMoves(Knight, color, 54, 42, WWS, Pawn)...)

				return moves
			},
		},
		{
			name: "jump over pieces",
			otherPieces: func() []Piece {
				return []Piece{
					NewPiece(Pawn, color, 74),
					NewPiece(Pawn, xColor, 64),
					NewPiece(Pawn, color, 44),
					NewPiece(Pawn, xColor, 34),
					NewPiece(Pawn, color, 55),
					NewPiece(Pawn, xColor, 56),
					NewPiece(Pawn, color, 53),
					NewPiece(Pawn, xColor, 52),
				}
			},
			expect: func() []Move {
				var moves []Move

				moves = append(moves, generateExpectedMoves(Knight, color, 54, 75, NNE)...)
				moves = append(moves, generateExpectedMoves(Knight, color, 54, 73, NNW)...)
				moves = append(moves, generateExpectedMoves(Knight, color, 54, 35, SSE)...)
				moves = append(moves, generateExpectedMoves(Knight, color, 54, 33, SSW)...)
				moves = append(moves, generateExpectedMoves(Knight, color, 54, 66, EEN)...)
				moves = append(moves, generateExpectedMoves(Knight, color, 54, 46, EES)...)
				moves = append(moves, generateExpectedMoves(Knight, color, 54, 62, WWN)...)
				moves = append(moves, generateExpectedMoves(Knight, color, 54, 42, WWS)...)

				return moves
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			board := NewEmptyBoard()
			tPiece := NewPiece(Knight, color, 54)
			err := board.LoadPieces(
				append(tc.otherPieces(), tPiece),
			)
			assert.NoError(t, err)
			moves, err := GeneratePiecePseudoLegalMoves(board, tPiece)
			assert.NoError(t, err)
			assert.Equal(t, tc.expect(), moves)
		})
	}
}
