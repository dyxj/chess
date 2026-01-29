package engine_test

import (
	"slices"
	"testing"

	. "github.com/dyxj/chess/internal/engine"
	"github.com/dyxj/chess/test/faker"
	"github.com/stretchr/testify/assert"
)

func TestKingPseudoLegalMoves(t *testing.T) {
	color := faker.Color()
	xColor := color.Opposite()

	tt := []struct {
		name        string
		piece       func() Piece
		otherPieces func() []Piece
		expect      func() []Move
	}{
		{
			name: "all directions",
			piece: func() Piece {
				return NewPiece(King, color, 54)
			},
			otherPieces: func() []Piece {
				return []Piece{}
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
			piece: func() Piece {
				return NewPiece(King, color, 54)
			},
			otherPieces: func() []Piece {
				return []Piece{
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
				return []Move{}
			},
		},
		{
			name: "capture",
			piece: func() Piece {
				return NewPiece(King, color, 54)
			},
			otherPieces: func() []Piece {
				return []Piece{
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
			tPiece := tc.piece()
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
				append([]Piece{}, tPiece),
			)
			assert.NoError(t, err)
			moves, err := GeneratePiecePseudoLegalMoves(board, tPiece)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedNumberOfMoves, len(moves))
		})
	}
}

func TestKingCastlingMoves(t *testing.T) {

	color := faker.Color()

	tt := []struct {
		name   string
		pieces func() []Piece
		expect func(kingPos int) []Move
	}{
		{
			name: "castling blocked by same color pieces",
			pieces: func() []Piece {
				var pieces []Piece
				pieces = GenerateStartPieces(color)
				pieces = slices.DeleteFunc(pieces, func(p Piece) bool {
					if p.Symbol() == Knight {
						return true
					}
					return false
				})
				return pieces
			},
			expect: func(kingPos int) []Move {
				return []Move{}
			},
		},
		{
			name: "castling blocked by opposite color pieces",
			pieces: func() []Piece {
				var pieces []Piece
				pieces = GenerateStartPieces(color)
				for i := 0; i < len(pieces); i++ {
					if pieces[i].Symbol() == Bishop {
						pieces[i] = NewPiece(Bishop, color.Opposite(), pieces[i].Position())
					}
				}
				pieces = slices.DeleteFunc(pieces, func(p Piece) bool {
					if p.Symbol() == Queen || p.Symbol() == Knight {
						return true
					}
					return false
				})

				return pieces
			},
			expect: func(kingPos int) []Move {
				var moves []Move

				moves = append(moves, Move{
					Color:    color,
					Symbol:   King,
					From:     kingPos,
					To:       kingPos + int(E),
					Captured: Bishop,
				})
				moves = append(moves, Move{
					Color:  color,
					Symbol: King,
					From:   kingPos,
					To:     kingPos + int(W),
				})

				return moves
			},
		},
		{
			name: "castling possible",
			pieces: func() []Piece {
				var pieces []Piece
				pieces = GenerateStartPieces(color)
				pieces = slices.DeleteFunc(pieces, func(p Piece) bool {
					if p.Symbol() == Queen || p.Symbol() == Knight || p.Symbol() == Bishop {
						return true
					}
					return false
				})

				return pieces
			},
			expect: func(kingPos int) []Move {
				var moves []Move
				base := 20
				if color == Black {
					base = 90
				}

				moves = append(moves, Move{
					Color:  color,
					Symbol: King,
					From:   kingPos,
					To:     kingPos + int(E),
				})
				moves = append(moves, Move{
					Color:  color,
					Symbol: King,
					From:   kingPos,
					To:     kingPos + int(W),
				})
				moves = append(moves, Move{
					Color:      color,
					Symbol:     King,
					From:       kingPos,
					To:         kingPos + int(W)*2,
					RookFrom:   base + 1,
					RookTo:     kingPos + int(W),
					IsCastling: true,
				})
				moves = append(moves, Move{
					Color:      color,
					Symbol:     King,
					From:       kingPos,
					To:         kingPos + int(E)*2,
					RookFrom:   base + 8,
					RookTo:     kingPos + int(E),
					IsCastling: true,
				})

				return moves
			},
		},
		{
			name: "castling not possible due to check",
			pieces: func() []Piece {
				var pieces []Piece
				pieces = GenerateStartPieces(color)
				pieces = slices.DeleteFunc(pieces, func(p Piece) bool {
					if p.Symbol() == Queen || p.Symbol() == Knight || p.Symbol() == Bishop {
						return true
					}
					if p.Symbol() == Pawn && p.Position()%10 == 5 {
						return true
					}
					return false
				})
				pieces = append(pieces, NewPiece(Rook, color.Opposite(), 55))

				return pieces
			},
			expect: func(kingPos int) []Move {
				var moves []Move

				direction := N
				if color == Black {
					direction = S
				}
				moves = append(moves, Move{
					Color:  color,
					Symbol: King,
					From:   kingPos,
					To:     kingPos + int(direction),
				})
				moves = append(moves, Move{
					Color:  color,
					Symbol: King,
					From:   kingPos,
					To:     kingPos + int(E),
				})
				moves = append(moves, Move{
					Color:  color,
					Symbol: King,
					From:   kingPos,
					To:     kingPos + int(W),
				})

				return moves
			},
		},
		{
			name: "castling not possible due to after castling check",
			pieces: func() []Piece {
				var pieces []Piece
				pieces = GenerateStartPieces(color)
				pieces = slices.DeleteFunc(pieces, func(p Piece) bool {
					if p.Symbol() == Queen || p.Symbol() == Knight || p.Symbol() == Bishop {
						return true
					}
					if p.Symbol() == Pawn &&
						(p.Position()%10 == 7 || p.Position()%10 == 3) {
						return true
					}
					return false
				})
				pieces = append(pieces, NewPiece(Rook, color.Opposite(), 53))
				pieces = append(pieces, NewPiece(Rook, color.Opposite(), 57))

				return pieces
			},
			expect: func(kingPos int) []Move {
				var moves []Move

				moves = append(moves, Move{
					Color:  color,
					Symbol: King,
					From:   kingPos,
					To:     kingPos + int(E),
				})
				moves = append(moves, Move{
					Color:  color,
					Symbol: King,
					From:   kingPos,
					To:     kingPos + int(W),
				})

				return moves
			},
		},
		{
			name: "castling not possible due to moved rook",
			pieces: func() []Piece {
				var pieces []Piece
				pieces = GenerateStartPieces(color)
				for i := 0; i < len(pieces); i++ {
					if pieces[i].Symbol() == Rook {
						pieces[i] = NewPiece(Rook, color, pieces[i].Position(), true)
					}
				}
				pieces = slices.DeleteFunc(pieces, func(p Piece) bool {
					if p.Symbol() == Queen || p.Symbol() == Knight || p.Symbol() == Bishop {
						return true
					}
					return false
				})

				return pieces
			},
			expect: func(kingPos int) []Move {
				var moves []Move

				moves = append(moves, Move{
					Color:  color,
					Symbol: King,
					From:   kingPos,
					To:     kingPos + int(E),
				})
				moves = append(moves, Move{
					Color:  color,
					Symbol: King,
					From:   kingPos,
					To:     kingPos + int(W),
				})

				return moves
			},
		},
		{
			name: "castling not possible due to moved king",
			pieces: func() []Piece {
				var pieces []Piece
				pieces = GenerateStartPieces(color)
				for i := 0; i < len(pieces); i++ {
					if pieces[i].Symbol() == King {
						pieces[i] = NewPiece(King, color, pieces[i].Position(), true)
					}
				}
				pieces = slices.DeleteFunc(pieces, func(p Piece) bool {
					if p.Symbol() == Queen || p.Symbol() == Knight || p.Symbol() == Bishop {
						return true
					}
					return false
				})

				return pieces
			},
			expect: func(kingPos int) []Move {
				var moves []Move

				moves = append(moves, Move{
					Color:  color,
					Symbol: King,
					From:   kingPos,
					To:     kingPos + int(E),
				})
				moves = append(moves, Move{
					Color:  color,
					Symbol: King,
					From:   kingPos,
					To:     kingPos + int(W),
				})

				return moves
			},
		},
		{
			name: "reached sentinel due to invalid board setup",
			pieces: func() []Piece {
				var pieces []Piece
				pieces = GenerateStartPieces(color)
				direction := N
				if color == Black {
					direction = S
				}
				for i := 0; i < len(pieces); i++ {
					if pieces[i].Symbol() == Rook {
						pieces[i] = NewPiece(Rook, color, pieces[i].Position()+int(direction)*2)
					}
				}
				pieces = slices.DeleteFunc(pieces, func(p Piece) bool {
					if p.Symbol() == Queen || p.Symbol() == Knight || p.Symbol() == Bishop {
						return true
					}
					return false
				})

				return pieces
			},
			expect: func(kingPos int) []Move {
				var moves []Move

				moves = append(moves, Move{
					Color:  color,
					Symbol: King,
					From:   kingPos,
					To:     kingPos + int(E),
				})
				moves = append(moves, Move{
					Color:  color,
					Symbol: King,
					From:   kingPos,
					To:     kingPos + int(W),
				})

				return moves
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			board := NewEmptyBoard()
			pieces := tc.pieces()
			err := board.LoadPieces(pieces)
			assert.NoError(t, err)
			kingIndex := slices.IndexFunc(pieces, func(p Piece) bool {
				if p.Symbol() == King {
					return true
				}
				return false
			})
			if !assert.NotEqual(t, -1, kingIndex) {
				t.FailNow()
			}
			king := pieces[kingIndex]
			moves, err := GeneratePiecePseudoLegalMoves(board, pieces[kingIndex])
			assert.NoError(t, err)
			moves = slices.DeleteFunc(moves, func(m Move) bool {
				if m.Symbol != King {
					return true
				}
				return false
			})
			assert.Equal(t, tc.expect(king.Position()), moves)
		})
	}
}
