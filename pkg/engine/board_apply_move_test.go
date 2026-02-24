package engine

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoard_ApplyMove_Validation(t *testing.T) {
	t.Run("not active color", func(t *testing.T) {
		board := NewBoard() // default white first
		err := board.ApplyMove(Move{
			Color:  Black,
			Symbol: Pawn,
			From:   83,
			To:     73,
		})
		assert.ErrorIs(t, err, ErrNotActiveColor)
	})

	t.Run("out of board(from)", func(t *testing.T) {
		board := NewBoard()
		err := board.ApplyMove(Move{
			Color:  White,
			Symbol: Pawn,
			From:   150,
			To:     73,
		})
		assert.ErrorIs(t, err, ErrOutOfBoard)
	})

	t.Run("out of board(to)", func(t *testing.T) {
		board := NewBoard()
		err := board.ApplyMove(Move{
			Color:  White,
			Symbol: Pawn,
			From:   83,
			To:     150,
		})
		assert.ErrorIs(t, err, ErrOutOfBoard)
	})

	t.Run("out of board(sentinel)", func(t *testing.T) {
		board := NewBoard()
		err := board.ApplyMove(Move{
			Color:  White,
			Symbol: Rook,
			From:   91,
			To:     90,
		})
		assert.ErrorIs(t, err, ErrOutOfBoard)
	})

	t.Run("piece not found(empty)", func(t *testing.T) {
		board := NewBoard()
		err := board.ApplyMove(Move{
			Color:  White,
			Symbol: Rook,
			From:   55,
			To:     45,
		})
		assert.ErrorIs(t, err, ErrPieceNotFound)
	})

	t.Run("piece not found(color)", func(t *testing.T) {
		board := NewBoard()
		err := board.ApplyMove(Move{
			Color:  White,
			Symbol: Pawn,
			From:   81,
			To:     71,
		})
		assert.ErrorIs(t, err, ErrPieceNotFound)
	})

	t.Run("piece not found(symbol)", func(t *testing.T) {
		board := NewBoard()
		err := board.ApplyMove(Move{
			Color:  White,
			Symbol: Rook,
			From:   31,
			To:     41,
		})
		assert.ErrorIs(t, err, ErrPieceNotFound)
	})
}

func TestBoard_ApplyMove_Normal(t *testing.T) {
	wPieces := GenerateStartPieces(White)

	board := NewEmptyBoard()

	err := board.LoadPieces(wPieces)
	assert.NoError(t, err)

	m := Move{
		Color:  White,
		Symbol: Pawn,
		From:   34,
		To:     45,
	}
	err = board.ApplyMove(m)
	assert.NoError(t, err)

	// position applied
	assert.Equal(t, EmptyCell, board.Value(34))
	assert.Equal(t, 1, board.Value(45))
	assert.Equal(t, 25, board.kingPosition(White))

	// piece list updated
	updatedPieceIndex := slices.IndexFunc(board.Pieces(White), func(p Piece) bool {
		if p.Symbol() == Pawn && p.Color() == White && p.Position() == 45 {
			return true
		}
		return false
	})
	assert.NotEqual(t, -1, updatedPieceIndex)

	// nothing placed in graveyard
	assert.Equal(t, 0, len(board.graveyard))

	// draw counter remains 0
	assert.Equal(t, 0, board.drawCounter)

	// active color flipped
	assert.Equal(t, Black, board.activeColor)

	// state hash incremented
	hash := board.calculateBoardStateHash(m, Black)
	hashCount, ok := board.boardStateHashMapCount[hash]
	assert.True(t, ok)
	assert.Equal(t, 1, hashCount)

	// round added to history
	lastRound, rOk := board.lastRound()
	assert.True(t, rOk)
	assert.Equal(t, round{
		Move:            m,
		PrevDrawCounter: 0,
		BoardStateHash:  hash,
	}, lastRound)
}

func TestBoard_ApplyMove_King(t *testing.T) {
	wPieces := GenerateStartPieces(White)
	// remove pawns
	pCount := 0
	for i, p := range wPieces {
		if p.Symbol() != Pawn {
			wPieces[pCount] = wPieces[i]
			pCount++
		}
	}
	wPieces = wPieces[:pCount]

	board := NewEmptyBoard()

	err := board.LoadPieces(wPieces)
	assert.NoError(t, err)

	m := Move{
		Color:  White,
		Symbol: King,
		From:   25,
		To:     35,
	}
	err = board.ApplyMove(m)
	assert.NoError(t, err)

	// position applied
	assert.Equal(t, EmptyCell, board.Value(34))
	assert.Equal(t, 6, board.Value(35))
	assert.Equal(t, 35, board.kingPosition(White))

	// piece list updated
	updatedPieceIndex := slices.IndexFunc(board.Pieces(White), func(p Piece) bool {
		if p.Symbol() == King && p.Color() == White && p.Position() == 35 {
			return true
		}
		return false
	})
	assert.NotEqual(t, -1, updatedPieceIndex)

	// nothing placed in graveyard
	assert.Equal(t, 0, len(board.graveyard))

	// draw counter remains 0
	assert.Equal(t, 1, board.drawCounter)

	// active color flipped
	assert.Equal(t, Black, board.activeColor)

	// state hash incremented
	hash := board.calculateBoardStateHash(m, Black)
	hashCount, ok := board.boardStateHashMapCount[hash]
	assert.True(t, ok)
	assert.Equal(t, 1, hashCount)

	// round added to history
	lastRound, rOk := board.lastRound()
	assert.True(t, rOk)
	assert.Equal(t, round{
		Move:            m,
		PrevDrawCounter: 0,
		BoardStateHash:  hash,
	}, lastRound)
}

func TestBoard_ApplyMove_Capture(t *testing.T) {
	wPieces := GenerateStartPieces(White)
	bPieces := GenerateStartPieces(Black)

	// remove pawns
	pCount := 0
	for i, p := range wPieces {
		if p.Symbol() != Pawn {
			wPieces[pCount] = wPieces[i]
			bPieces[pCount] = bPieces[i]
			pCount++
		}
	}
	wPieces = wPieces[:pCount]
	bPieces = bPieces[:pCount]

	board := NewEmptyBoard()

	err := board.LoadPieces(wPieces)
	assert.NoError(t, err)
	err = board.LoadPieces(bPieces)
	assert.NoError(t, err)

	m := Move{
		Color:    White,
		Symbol:   Rook,
		From:     21,
		To:       91,
		Captured: Rook,
	}
	err = board.ApplyMove(m)
	assert.NoError(t, err)

	// position applied
	assert.Equal(t, EmptyCell, board.Value(21))
	assert.Equal(t, 4, board.Value(91))

	// white piece list updated
	updatedPieceIndex := slices.IndexFunc(board.Pieces(White), func(p Piece) bool {
		if p.Symbol() == Rook && p.Color() == White && p.Position() == 91 {
			return true
		}
		return false
	})
	assert.NotEqual(t, -1, updatedPieceIndex)

	// black piece list updated
	capturedIndex := slices.IndexFunc(board.Pieces(Black), func(p Piece) bool {
		if p.Symbol() == Rook && p.Color() == Black && p.Position() == 91 {
			return true
		}
		return false
	})
	// should not be in black piece list
	assert.Equal(t, -1, capturedIndex)

	graveyardIndex := slices.IndexFunc(board.graveyard, func(p Piece) bool {
		if p.Symbol() == Rook && p.Color() == Black && p.Position() == 91 {
			return true
		}
		return false
	})
	// should find black rook in graveyard
	assert.NotEqual(t, -1, graveyardIndex)

	// draw counter remains 0
	assert.Equal(t, 0, board.drawCounter)

	// active color flipped
	assert.Equal(t, Black, board.activeColor)

	// state hash incremented
	hash := board.calculateBoardStateHash(m, Black)
	hashCount, ok := board.boardStateHashMapCount[hash]
	assert.True(t, ok)
	assert.Equal(t, 1, hashCount)

	// round added to history
	lastRound, rOk := board.lastRound()
	assert.True(t, rOk)
	assert.Equal(t, round{
		Move:            m,
		PrevDrawCounter: 0,
		BoardStateHash:  hash,
	}, lastRound)
}

func TestBoard_ApplyMove_Capture_With_Promotion(t *testing.T) {

	board := NewEmptyBoard()
	from := 81
	to := 92
	pieces := []Piece{
		NewPiece(Pawn, White, from),
		NewPiece(Knight, Black, to),
	}
	err := board.LoadPieces(pieces)
	assert.NoError(t, err)

	m := Move{
		Color:     White,
		Symbol:    Pawn,
		From:      from,
		To:        to,
		Captured:  Knight,
		Promotion: Queen,
	}
	err = board.ApplyMove(m)
	assert.NoError(t, err)

	// position applied
	assert.Equal(t, EmptyCell, board.Value(from))
	assert.Equal(t, 5, board.Value(to))

	// white piece list updated pawn
	updatedPieceIndex := slices.IndexFunc(board.Pieces(White), func(p Piece) bool {
		if p.Symbol() == Pawn && p.Color() == White &&
			(p.Position() == to || p.Position() == from) {
			return true
		}
		return false
	})
	// pawn not found in list
	assert.Equal(t, -1, updatedPieceIndex)

	promotedIndex := slices.IndexFunc(board.Pieces(White), func(p Piece) bool {
		if p.Symbol() == Queen && p.Color() == White && p.Position() == to {
			return true
		}
		return false
	})
	// found queen in new list
	assert.NotEqual(t, -1, promotedIndex)

	// black piece list updated
	capturedIndex := slices.IndexFunc(board.Pieces(Black), func(p Piece) bool {
		if p.Symbol() == Knight && p.Color() == Black && p.Position() == to {
			return true
		}
		return false
	})
	// should not be in black piece list
	assert.Equal(t, -1, capturedIndex)

	graveyardIndex := slices.IndexFunc(board.graveyard, func(p Piece) bool {
		if p.Symbol() == Knight && p.Color() == Black && p.Position() == to {
			return true
		}
		return false
	})
	// should find black piece in graveyard
	assert.NotEqual(t, -1, graveyardIndex)

	// draw counter remains 0
	assert.Equal(t, 0, board.drawCounter)

	// active color flipped
	assert.Equal(t, Black, board.activeColor)

	// state hash incremented
	hash := board.calculateBoardStateHash(m, Black)
	hashCount, ok := board.boardStateHashMapCount[hash]
	assert.True(t, ok)
	assert.Equal(t, 1, hashCount)

	// round added to history
	lastRound, rOk := board.lastRound()
	assert.True(t, rOk)
	assert.Equal(t, round{
		Move:            m,
		PrevDrawCounter: 0,
		BoardStateHash:  hash,
	}, lastRound)
}

func TestBoard_ApplyMove_EnPassant(t *testing.T) {
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
		To:          wPawnPos + int(NW),
		Captured:    Pawn,
		IsEnPassant: true,
	}
	err = board.ApplyMove(enPassantMove)
	assert.NoError(t, err)

	// position applied
	assert.Equal(t, EmptyCell, board.Value(wPawnPos))      // original white pawn pos is empty
	assert.Equal(t, EmptyCell, board.Value(bPawnEndPos))   // captured black pawn pos is empty
	assert.Equal(t, int(Pawn)*int(White), board.Value(74)) // new white pawn pos

	// white piece list updated
	updatedPieceIndex := slices.IndexFunc(board.Pieces(White), func(p Piece) bool {
		return p.Symbol() == Pawn && p.Position() == 74
	})
	assert.NotEqual(t, -1, updatedPieceIndex, "White pawn should be at new position")

	// black piece list updated
	capturedIndex := slices.IndexFunc(board.Pieces(Black), func(p Piece) bool {
		return p.Position() == bPawnEndPos
	})
	assert.Equal(t, -1, capturedIndex, "Captured black pawn should be removed from piece list")

	// graveyard updated
	graveyardIndex := slices.IndexFunc(board.graveyard, func(p Piece) bool {
		return p.Symbol() == Pawn && p.Color() == Black && p.Position() == bPawnEndPos
	})
	assert.NotEqual(t, -1, graveyardIndex, "Captured black pawn should be in graveyard")

	// draw counter reset
	assert.Equal(t, 0, board.drawCounter)

	// active color flipped(black > white > black)
	assert.Equal(t, Black, board.activeColor)

	// state hash incremented
	hash := board.calculateBoardStateHash(enPassantMove, Black)
	hashCount, ok := board.boardStateHashMapCount[hash]
	assert.True(t, ok)
	assert.Equal(t, 1, hashCount)

	// round added to history
	lastRound, rOk := board.lastRound()
	assert.True(t, rOk)
	assert.Equal(t, round{
		Move:            enPassantMove,
		PrevDrawCounter: 0,
		BoardStateHash:  hash,
	}, lastRound)
}

func TestBoard_ApplyMove_Castling(t *testing.T) {
	board := NewEmptyBoard()
	pieces := []Piece{
		NewPiece(King, White, 25),
		NewPiece(Rook, White, 28),
	}
	err := board.LoadPieces(pieces)
	assert.NoError(t, err)

	// White performs king-side castling
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

	// position applied
	assert.Equal(t, EmptyCell, board.Value(25), "original king pos is empty")
	assert.Equal(t, EmptyCell, board.Value(28), "original rook pos is empty")
	assert.Equal(t, int(King)*int(White), board.Value(27), "new king pos")
	assert.Equal(t, int(Rook)*int(White), board.Value(26), "new rook pos")

	// white piece list updated
	kingIndex := slices.IndexFunc(board.Pieces(White), func(p Piece) bool {
		return p.Symbol() == King && p.Position() == 27
	})
	assert.NotEqual(t, -1, kingIndex, "King should be at new position")

	rookIndex := slices.IndexFunc(board.Pieces(White), func(p Piece) bool {
		return p.Symbol() == Rook && p.Position() == 26
	})
	assert.NotEqual(t, -1, rookIndex, "Rook should be at new position")

	// draw counter incremented
	assert.Equal(t, 1, board.drawCounter)

	// active color flipped
	assert.Equal(t, Black, board.activeColor)

	// state hash incremented
	hash := board.calculateBoardStateHash(castlingMove, Black)
	hashCount, ok := board.boardStateHashMapCount[hash]
	assert.True(t, ok)
	assert.Equal(t, 1, hashCount)

	// round added to history
	lastRound, rOk := board.lastRound()
	assert.True(t, rOk)
	assert.Equal(t, round{
		Move:            castlingMove,
		PrevDrawCounter: 0,
		BoardStateHash:  hash,
	}, lastRound)
}
