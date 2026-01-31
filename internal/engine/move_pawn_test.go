package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPawnPseudoLegalMoves(t *testing.T) {

	tt := []struct {
		name          string
		startingPiece func() Piece
		otherPieces   func() []Piece
		expect        func() []Move
	}{
		{
			name: "start position(white)",
			startingPiece: func() Piece {
				return NewPiece(Pawn, White, 34)
			},
			otherPieces: func() []Piece {
				var pieces []Piece
				return pieces
			},
			expect: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 34, To: 44})
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 34, To: 54})
				return moves
			},
		},
		{
			name: "start position(black)",
			startingPiece: func() Piece {
				return NewPiece(Pawn, Black, 84)
			},
			otherPieces: func() []Piece {
				var pieces []Piece
				return pieces
			},
			expect: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 84, To: 74})
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 84, To: 64})
				return moves
			},
		},
		{
			name: "start position blocked 1(white)",
			startingPiece: func() Piece {
				return NewPiece(Pawn, White, 34)
			},
			otherPieces: func() []Piece {
				var pieces []Piece
				pieces = append(pieces, NewPiece(Pawn, Black, 44))
				return pieces
			},
			expect: func() []Move {
				return []Move{}
			},
		},
		{
			name: "start position blocked 1(black)",
			startingPiece: func() Piece {
				return NewPiece(Pawn, Black, 84)
			},
			otherPieces: func() []Piece {
				var pieces []Piece
				pieces = append(pieces, NewPiece(Pawn, White, 74))
				return pieces
			},
			expect: func() []Move {
				return []Move{}
			},
		},
		{
			name: "start position blocked 2(white)",
			startingPiece: func() Piece {
				return NewPiece(Pawn, White, 34)
			},
			otherPieces: func() []Piece {
				var pieces []Piece
				pieces = append(pieces, NewPiece(Pawn, Black, 54))
				return pieces
			},
			expect: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 34, To: 44})
				return moves
			},
		},
		{
			name: "start position blocked 2(black)",
			startingPiece: func() Piece {
				return NewPiece(Pawn, Black, 84)
			},
			otherPieces: func() []Piece {
				var pieces []Piece
				pieces = append(pieces, NewPiece(Pawn, White, 64))
				return pieces
			},
			expect: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 84, To: 74})
				return moves
			},
		},
		{
			name: "start position with capture(white)",
			startingPiece: func() Piece {
				return NewPiece(Pawn, White, 34)
			},
			otherPieces: func() []Piece {
				var pieces []Piece
				pieces = append(pieces, NewPiece(Pawn, Black, 45))
				pieces = append(pieces, NewPiece(Pawn, Black, 43))
				return pieces
			},
			expect: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 34, To: 44})
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 34, To: 54})
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 34, To: 45, Captured: Pawn})
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 34, To: 43, Captured: Pawn})
				return moves
			},
		},
		{
			name: "start position with capture(black)",
			startingPiece: func() Piece {
				return NewPiece(Pawn, Black, 84)
			},
			otherPieces: func() []Piece {
				var pieces []Piece
				pieces = append(pieces, NewPiece(Pawn, White, 75))
				pieces = append(pieces, NewPiece(Pawn, White, 73))
				return pieces
			},
			expect: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 84, To: 74})
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 84, To: 64})
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 84, To: 75, Captured: Pawn})
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 84, To: 73, Captured: Pawn})
				return moves
			},
		},
		{
			name: "moved position with capture(white)",
			startingPiece: func() Piece {
				piece := NewPiece(Pawn, White, 44)
				piece.moveCount = 1
				return piece
			},
			otherPieces: func() []Piece {
				var pieces []Piece
				pieces = append(pieces, NewPiece(Pawn, Black, 55))
				pieces = append(pieces, NewPiece(Pawn, Black, 53))
				return pieces
			},
			expect: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 44, To: 54})
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 44, To: 55, Captured: Pawn})
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 44, To: 53, Captured: Pawn})
				return moves
			},
		},
		{
			name: "moved position with capture(black)",
			startingPiece: func() Piece {
				piece := NewPiece(Pawn, Black, 74)
				piece.moveCount = 1
				return piece
			},
			otherPieces: func() []Piece {
				var pieces []Piece
				pieces = append(pieces, NewPiece(Pawn, White, 65))
				pieces = append(pieces, NewPiece(Pawn, White, 63))
				return pieces
			},
			expect: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 74, To: 64})
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 74, To: 65, Captured: Pawn})
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 74, To: 63, Captured: Pawn})
				return moves
			},
		},
		{
			name: "promotion(white)",
			startingPiece: func() Piece {
				piece := NewPiece(Pawn, White, 84)
				piece.moveCount = 1
				return piece
			},
			otherPieces: func() []Piece {
				var pieces []Piece
				return pieces
			},
			expect: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 84, To: 94, Promotion: Queen})
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 84, To: 94, Promotion: Rook})
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 84, To: 94, Promotion: Bishop})
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 84, To: 94, Promotion: Knight})
				return moves
			},
		},
		{
			name: "promotion(black)",
			startingPiece: func() Piece {
				piece := NewPiece(Pawn, Black, 34)
				piece.moveCount = 1
				return piece
			},
			otherPieces: func() []Piece {
				var pieces []Piece
				return pieces
			},
			expect: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 34, To: 24, Promotion: Queen})
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 34, To: 24, Promotion: Rook})
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 34, To: 24, Promotion: Bishop})
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 34, To: 24, Promotion: Knight})
				return moves
			},
		},
		{
			name: "capture promotion(white)",
			startingPiece: func() Piece {
				piece := NewPiece(Pawn, White, 84)
				piece.moveCount = 1
				return piece
			},
			otherPieces: func() []Piece {
				var pieces []Piece
				pieces = append(pieces, NewPiece(Rook, Black, 94))
				pieces = append(pieces, NewPiece(Rook, Black, 95))
				pieces = append(pieces, NewPiece(Rook, Black, 93))
				return pieces
			},
			expect: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 84, To: 95, Promotion: Queen, Captured: Rook})
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 84, To: 95, Promotion: Rook, Captured: Rook})
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 84, To: 95, Promotion: Bishop, Captured: Rook})
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 84, To: 95, Promotion: Knight, Captured: Rook})
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 84, To: 93, Promotion: Queen, Captured: Rook})
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 84, To: 93, Promotion: Rook, Captured: Rook})
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 84, To: 93, Promotion: Bishop, Captured: Rook})
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 84, To: 93, Promotion: Knight, Captured: Rook})
				return moves
			},
		},
		{
			name: "capture promotion(black)",
			startingPiece: func() Piece {
				piece := NewPiece(Pawn, Black, 34)
				piece.moveCount = 1
				return piece
			},
			otherPieces: func() []Piece {
				var pieces []Piece
				pieces = append(pieces, NewPiece(Rook, White, 24))
				pieces = append(pieces, NewPiece(Rook, White, 25))
				pieces = append(pieces, NewPiece(Rook, White, 23))
				return pieces
			},
			expect: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 34, To: 25, Promotion: Queen, Captured: Rook})
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 34, To: 25, Promotion: Rook, Captured: Rook})
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 34, To: 25, Promotion: Bishop, Captured: Rook})
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 34, To: 25, Promotion: Knight, Captured: Rook})
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 34, To: 23, Promotion: Queen, Captured: Rook})
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 34, To: 23, Promotion: Rook, Captured: Rook})
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 34, To: 23, Promotion: Bishop, Captured: Rook})
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 34, To: 23, Promotion: Knight, Captured: Rook})
				return moves
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			board := NewEmptyBoard()
			tPiece := tc.startingPiece()
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

func TestPawnEnPassantPseudoLegalMoves(t *testing.T) {
	tt := []struct {
		name          string
		startingPiece func() Piece
		otherPieces   func() []Piece
		moveHistory   func() []Move
		expect        func() []Move
	}{
		{
			name: "white(east)",
			startingPiece: func() Piece {
				piece := NewPiece(Pawn, White, 64)
				piece.moveCount = 1
				return piece
			},
			otherPieces: func() []Piece {
				var pieces []Piece
				piece := NewPiece(Pawn, Black, 65)
				piece.moveCount = 1
				pieces = append(pieces, piece)
				return pieces
			},
			moveHistory: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 85, To: 65})
				return moves
			},
			expect: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 64, To: 74})
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 64, To: 75, IsEnPassant: true, Captured: Pawn})
				return moves
			},
		},
		{
			name: "white(west)",
			startingPiece: func() Piece {
				piece := NewPiece(Pawn, White, 64)
				piece.moveCount = 1
				return piece
			},
			otherPieces: func() []Piece {
				var pieces []Piece
				piece := NewPiece(Pawn, Black, 63)
				piece.moveCount = 1
				pieces = append(pieces, piece)
				return pieces
			},
			moveHistory: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 83, To: 63})
				return moves
			},
			expect: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 64, To: 74})
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 64, To: 73, IsEnPassant: true, Captured: Pawn})
				return moves
			},
		},
		{
			name: "black(east)",
			startingPiece: func() Piece {
				piece := NewPiece(Pawn, Black, 54)
				piece.moveCount = 1
				return piece
			},
			otherPieces: func() []Piece {
				var pieces []Piece
				piece := NewPiece(Pawn, White, 55)
				piece.moveCount = 1
				pieces = append(pieces, piece)
				return pieces
			},
			moveHistory: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 35, To: 55})
				return moves
			},
			expect: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 54, To: 44})
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 54, To: 45, IsEnPassant: true, Captured: Pawn})
				return moves
			},
		},
		{
			name: "black(west)",
			startingPiece: func() Piece {
				piece := NewPiece(Pawn, Black, 54)
				piece.moveCount = 1
				return piece
			},
			otherPieces: func() []Piece {
				var pieces []Piece
				piece := NewPiece(Pawn, White, 53)
				piece.moveCount = 1
				pieces = append(pieces, piece)
				return pieces
			},
			moveHistory: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 33, To: 53})
				return moves
			},
			expect: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 54, To: 44})
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 54, To: 43, IsEnPassant: true, Captured: Pawn})
				return moves
			},
		},
		{
			name: "last move not pawn double step, but is beside pawn that double stepped earlier",
			startingPiece: func() Piece {
				piece := NewPiece(Pawn, White, 64)
				piece.moveCount = 1
				return piece
			},
			otherPieces: func() []Piece {
				var pieces []Piece
				piece := NewPiece(Pawn, Black, 63)
				piece.moveCount = 1
				pieces = append(pieces, piece, NewPiece(Rook, Black, 41), NewPiece(Rook, Black, 78))
				return pieces
			},
			moveHistory: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 83, To: 63})
				moves = append(moves, Move{Color: White, Symbol: Rook, From: 98, To: 78})
				moves = append(moves, Move{Color: Black, Symbol: Rook, From: 21, To: 41})
				return moves
			},
			expect: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 64, To: 74})
				return moves
			},
		},
		{
			name: "move history and board out of sync",
			startingPiece: func() Piece {
				piece := NewPiece(Pawn, White, 64)
				piece.moveCount = 1
				return piece
			},
			otherPieces: func() []Piece {
				var pieces []Piece
				// Note that this should be a Pawn
				piece := NewPiece(Rook, Black, 63)
				piece.moveCount = 1
				pieces = append(pieces, piece)
				return pieces
			},
			moveHistory: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: Black, Symbol: Pawn, From: 83, To: 63})
				return moves
			},
			expect: func() []Move {
				var moves []Move
				moves = append(moves, Move{Color: White, Symbol: Pawn, From: 64, To: 74})
				return moves
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			board := NewEmptyBoard()
			board.moveHistory = tc.moveHistory()
			tPiece := tc.startingPiece()
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
