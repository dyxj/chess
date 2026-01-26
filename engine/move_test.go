package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRookBasicMoves(t *testing.T) {
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

				moves = append(moves, generateExpectedMoves(White, Rook, 54, 94, N)...)
				moves = append(moves, generateExpectedMoves(White, Rook, 54, 24, S)...)
				moves = append(moves, generateExpectedMoves(White, Rook, 54, 58, E)...)
				moves = append(moves, generateExpectedMoves(White, Rook, 54, 51, W)...)

				return moves
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			board := NewEmptyBoard()
			tPiece := NewPiece(Rook, White, 54)
			err := board.LoadPieces(
				append(tc.otherPieces(), tPiece),
			)
			assert.NoError(t, err)
			moves := generateBasicMoves(board, tPiece)
			assert.Equal(t, tc.expect(), moves)
		})
	}
}

// generateExpectedMoves generates moves from [from,to] in the given direction.
func generateExpectedMoves(color Color, symbol Symbol, from int, to int, direction Direction) []Move {
	moves := make([]Move, 0)

	i := from
	shouldGen := true
	for shouldGen {
		moves = append(moves, Move{
			Color:  color,
			Symbol: symbol,
			From:   i,
			To:     i + int(direction),
		})
		i += int(direction)
		if direction > 0 {
			if i >= to {
				shouldGen = false
			}
		} else {
			if i <= to {
				shouldGen = false
			}
		}
	}
	return moves
}
