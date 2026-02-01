package engine

import (
	"testing"

	"github.com/dyxj/chess/pkg/randx"
	"github.com/stretchr/testify/assert"
)

func TestGeneratePseudoLegalMovesErrors(t *testing.T) {
	color := randx.FromSlice(Colors)
	tt := []struct {
		name     string
		selected func() Piece
		pieces   func() []Piece
		expected error
	}{
		{
			name: "piece not found(symbol mismatch)",
			selected: func() Piece {
				return NewPiece(Rook, color, 74)
			},
			pieces: func() []Piece {
				var pieces []Piece
				pieces = append(pieces, NewPiece(Pawn, color, 74))
				return pieces
			},
			expected: ErrPieceNotFound,
		},
		{
			name: "piece not found(color mismatch)",
			selected: func() Piece {
				return NewPiece(Pawn, color.Opposite(), 74)
			},
			pieces: func() []Piece {
				var pieces []Piece
				pieces = append(pieces, NewPiece(Pawn, color, 74))
				return pieces
			},
			expected: ErrPieceNotFound,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			board := NewEmptyBoard()
			err := board.LoadPieces(tc.pieces())
			assert.NoError(t, err)
			_, err = GeneratePiecePseudoLegalMoves(board, tc.selected())
			assert.Equal(t, tc.expected, err)
		})
	}
}

// generateExpectedMoves generates moves from [from,to] in the given direction.
// takes first capture and assigns it to last move.
func generateExpectedMoves(symbol Symbol, color Color, from int, to int, direction Direction, capture ...Symbol) []Move {
	moves := make([]Move, 0)

	i := from
	for {
		if direction > 0 {
			if i >= to {
				break
			}
		} else {
			if i <= to {
				break
			}
		}
		moves = append(moves, Move{
			Color:  color,
			Symbol: symbol,
			From:   from,
			To:     i + int(direction),
		})
		i += int(direction)
	}

	if len(capture) > 0 && len(moves) > 0 {
		moves[len(moves)-1].Captured = capture[0]
	}
	return moves
}
