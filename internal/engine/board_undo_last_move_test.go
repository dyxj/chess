package engine

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoard_UndoLastMove_NoMove(t *testing.T) {
	b := NewEmptyBoard()
	ok := b.UndoLastMove()
	assert.False(t, ok)
}

func TestBoard_UndoLastMove(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		board := NewBoard()
		m := Move{
			Color:  White,
			Symbol: Pawn,
			From:   34,
			To:     54,
		}
		err := board.ApplyMove(m)
		assert.NoError(t, err)
		hash := board.calculateBoardStateHash(m, Black)

		ok := board.UndoLastMove()
		assert.True(t, ok)

		// position restored
		assert.Equal(t, int(Pawn)*int(White), board.Value(34))
		assert.Equal(t, EmptyCell, board.Value(54))

		// piece list restored
		p, ok := board.Piece(White, Pawn, 34)
		assert.True(t, ok)
		assert.Equal(t, 0, p.moveCount)

		// graveyard empty
		assert.Empty(t, board.graveyard)

		// draw counter restored
		assert.Equal(t, 0, board.drawCounter)

		// active color restored
		assert.Equal(t, White, board.activeColor)

		// state hash decremented
		hashCount, ok := board.boardStateHashMapCount[hash]
		assert.True(t, ok)
		assert.Equal(t, 0, hashCount)

		// round removed from history
		_, rOk := board.lastRound()
		assert.False(t, rOk)
	})

	t.Run("capture", func(t *testing.T) {
		board := NewBoard()
		m := Move{
			Color:    White,
			Symbol:   Rook,
			From:     21,
			To:       81,
			Captured: Pawn,
		}
		err := board.ApplyMove(m)
		assert.NoError(t, err)

		ok := board.UndoLastMove()
		assert.True(t, ok)

		// position restored
		assert.Equal(t, int(Rook)*int(White), board.Value(21))
		assert.Equal(t, int(Pawn)*int(Black), board.Value(81))

		// piece lists restored
		_, ok = board.Piece(White, Rook, 21)
		assert.True(t, ok)
		_, ok = board.Piece(Black, Pawn, 81)
		assert.True(t, ok)

		// graveyard empty
		assert.Empty(t, board.graveyard)
	})

	t.Run("capture with promotion", func(t *testing.T) {
		board := NewEmptyBoard()
		from := 81
		to := 91
		pieces := []Piece{
			NewPiece(Pawn, White, from),
			NewPiece(Rook, Black, to),
		}
		err := board.LoadPieces(pieces)
		assert.NoError(t, err)

		m := Move{
			Color:     White,
			Symbol:    Pawn,
			From:      from,
			To:        to,
			Captured:  Rook,
			Promotion: Queen,
		}
		err = board.ApplyMove(m)
		assert.NoError(t, err)

		ok := board.UndoLastMove()
		assert.True(t, ok)

		// position restored
		assert.Equal(t, int(Pawn)*int(White), board.Value(from))
		assert.Equal(t, int(Rook)*int(Black), board.Value(to))

		// piece lists restored
		_, ok = board.Piece(White, Pawn, from)
		assert.True(t, ok)
		_, ok = board.Piece(Black, Rook, to)
		assert.True(t, ok)
		_, ok = board.Piece(White, Queen, to)
		assert.False(t, ok)
	})

	t.Run("en passant", func(t *testing.T) {
		board := NewEmptyBoard(Black)
		wPawnPos := 65
		bPawnStartPos := 84
		bPawnEndPos := 64
		pieces := []Piece{
			NewPiece(Pawn, White, wPawnPos),
			NewPiece(Pawn, Black, bPawnStartPos),
		}
		err := board.LoadPieces(pieces)
		assert.NoError(t, err)

		// Black moves pawn two squares
		blackMove := Move{
			Color:  Black,
			Symbol: Pawn,
			From:   bPawnStartPos,
			To:     bPawnEndPos,
		}
		err = board.ApplyMove(blackMove)
		assert.NoError(t, err)

		// White performs en passant
		enPassantMove := Move{
			Color:       White,
			Symbol:      Pawn,
			From:        wPawnPos,
			To:          74,
			Captured:    Pawn,
			IsEnPassant: true,
		}
		err = board.ApplyMove(enPassantMove)
		assert.NoError(t, err)

		ok := board.UndoLastMove()
		assert.True(t, ok)

		// Positions restored
		assert.Equal(t, int(Pawn)*int(White), board.Value(wPawnPos))
		assert.Equal(t, int(Pawn)*int(Black), board.Value(bPawnEndPos))
		assert.Equal(t, EmptyCell, board.Value(74))

		// Piece lists restored
		_, ok = board.Piece(White, Pawn, wPawnPos)
		assert.True(t, ok)
		_, ok = board.Piece(Black, Pawn, bPawnEndPos)
		assert.True(t, ok)

		// Active color restored
		assert.Equal(t, White, board.activeColor)
	})

	t.Run("castling", func(t *testing.T) {
		board := NewEmptyBoard()
		pieces := []Piece{
			NewPiece(King, White, 25),
			NewPiece(Rook, White, 28),
		}
		err := board.LoadPieces(pieces)
		assert.NoError(t, err)

		castlingMove := Move{
			Color:      White,
			Symbol:     King,
			From:       25,
			To:         27,
			RookFrom:   28,
			RookTo:     26,
			IsCastling: true,
		}
		err = board.ApplyMove(castlingMove)
		assert.NoError(t, err)

		ok := board.UndoLastMove()
		assert.True(t, ok)

		// Positions restored
		assert.Equal(t, int(King)*int(White), board.Value(25))
		assert.Equal(t, int(Rook)*int(White), board.Value(28))
		assert.Equal(t, EmptyCell, board.Value(27))
		assert.Equal(t, EmptyCell, board.Value(26))

		// Piece lists restored
		k, ok := board.Piece(White, King, 25)
		assert.True(t, ok)
		assert.Equal(t, 0, k.moveCount)
		r, ok := board.Piece(White, Rook, 28)
		assert.True(t, ok)
		assert.Equal(t, 0, r.moveCount)

		// King position restored
		assert.Equal(t, 25, board.kingPosition(White))
	})

	t.Run("capture(panic: invalid graveyard symbol)", func(t *testing.T) {
		assert.PanicsWithValue(t, "last graveyard symbol does not match move symbol",
			func() {
				board := NewBoard()
				m := Move{
					Color:    White,
					Symbol:   Rook,
					From:     21,
					To:       81,
					Captured: Pawn,
				}
				err := board.ApplyMove(m)
				assert.NoError(t, err)

				board.graveyard[0].symbol = Queen

				_ = board.UndoLastMove()
			})
	})

	t.Run("capture(panic: empty graveyard)", func(t *testing.T) {
		assert.PanicsWithValue(t, "expect piece in graveyard but it is empty", func() {
			board := NewBoard()
			m := Move{
				Color:    White,
				Symbol:   Rook,
				From:     21,
				To:       81,
				Captured: Pawn,
			}
			err := board.ApplyMove(m)
			assert.NoError(t, err)
			board.popGraveyard()
			fmt.Println(board.graveyard)

			_ = board.UndoLastMove()
		})
	})
}
