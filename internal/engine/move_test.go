package engine

import (
	"math/rand/v2"
	"slices"
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
			_, err = board.GeneratePiecePseudoLegalMoves(tc.selected())
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestMoveMethods(t *testing.T) {
	m := Move{
		Color:       randx.FromSlice(Colors),
		Symbol:      randx.FromSlice(Symbols),
		From:        rand.IntN(boardSize),
		To:          rand.IntN(boardSize),
		IsCastling:  randx.Bool(),
		RookFrom:    rand.IntN(boardSize),
		RookTo:      rand.IntN(boardSize),
		Captured:    randx.FromSlice(Symbols),
		Promotion:   randx.FromSlice(Symbols),
		IsEnPassant: randx.Bool(),
	}

	m2 := Move{}

	assert.Equal(t, true, m.hasCaptured())
	assert.Equal(t, false, m2.hasCaptured())

	assert.Equal(t, true, m.hasPromotion())
	assert.Equal(t, false, m2.hasPromotion())

	direction := N
	if m.Color == Black {
		direction = S
	}
	assert.Equal(t, m.To-int(direction), m.calculateEnPassantCapturedPos())
}

func TestHasLegalMoves_True(t *testing.T) {
	b := NewEmptyBoard(Black)
	err := b.LoadPieces(
		[]Piece{
			NewPiece(Rook, White, 91),
			NewPiece(Rook, White, 81),
			NewPiece(King, Black, 83),
		},
	)
	assert.NoError(t, err)

	hasLegalMoves := b.HasLegalMoves(b.ActiveColor())
	assert.True(t, hasLegalMoves)
}

func TestHasLegalMoves_False(t *testing.T) {
	b := NewEmptyBoard(Black)
	err := b.LoadPieces(
		[]Piece{
			NewPiece(Rook, White, 91),
			NewPiece(Rook, White, 81),
			NewPiece(King, Black, 93),
		},
	)
	assert.NoError(t, err)

	hasLegalMoves := b.HasLegalMoves(b.ActiveColor())
	assert.False(t, hasLegalMoves)
}

func TestHasLegalMoves_Panic(t *testing.T) {
	assert.Panics(t, func() {
		b := NewEmptyBoard(Black)
		err := b.LoadPieces(
			[]Piece{
				NewPiece(Rook, White, 91),
				NewPiece(Rook, White, 81),
				NewPiece(King, Black, 93),
			},
		)
		assert.NoError(t, err)

		b.cells[93] = EmptyCell
		hasLegalMoves := b.HasLegalMoves(b.ActiveColor())
		assert.False(t, hasLegalMoves)
	})
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

func sortMoves(moves []Move) {
	slices.SortStableFunc(moves, func(a, b Move) int {
		return moveSortingValue(a) - moveSortingValue(b)
	})
}

func moveSortingValue(m Move) int {
	return int(m.Symbol) * (m.From + m.To)
}
