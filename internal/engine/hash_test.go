package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateBoardStateHash(t *testing.T) {
	t.Run("same position produces same hash", func(t *testing.T) {
		board1 := NewBoard()
		board2 := NewBoard()

		hash1 := board1.calculateBoardStateHash(Move{}, White)
		hash2 := board2.calculateBoardStateHash(Move{}, White)

		assert.Equal(t, hash1, hash2)
	})

	t.Run("different piece positions produce different hashes", func(t *testing.T) {
		board1 := NewEmptyBoard()
		board2 := NewEmptyBoard()

		_ = board1.LoadPieces([]Piece{
			NewPiece(Pawn, White, 32, false),
			NewPiece(King, White, 25, false),
			NewPiece(King, Black, 95, false),
		})

		_ = board2.LoadPieces([]Piece{
			NewPiece(Pawn, White, 52, false),
			NewPiece(King, White, 25, false),
			NewPiece(King, Black, 95, false),
		})

		hash1 := board1.calculateBoardStateHash(Move{}, White)
		hash2 := board2.calculateBoardStateHash(Move{}, White)

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("different active color produces different hash", func(t *testing.T) {
		board := NewBoard()

		hashWhite := board.calculateBoardStateHash(Move{}, White)
		hashBlack := board.calculateBoardStateHash(Move{}, Black)

		assert.NotEqualf(t, hashWhite, hashBlack, "%v %v", hashWhite, hashBlack)
	})

	t.Run("different castling rights produce different hashes", func(t *testing.T) {
		// Board with castling available
		board1 := NewEmptyBoard()
		_ = board1.LoadPieces([]Piece{
			NewPiece(King, White, 25, false), // Unmoved king
			NewPiece(Rook, White, 21, false), // Unmoved west rook
			NewPiece(Rook, White, 28, false), // Unmoved east rook
			NewPiece(King, Black, 95, false),
		})

		// Board with king moved (no castling)
		board2 := NewEmptyBoard()
		_ = board2.LoadPieces([]Piece{
			NewPiece(King, White, 25, true), // Moved king
			NewPiece(Rook, White, 21, false),
			NewPiece(Rook, White, 28, false),
			NewPiece(King, Black, 95, false),
		})

		hash1 := board1.calculateBoardStateHash(Move{}, White)
		hash2 := board2.calculateBoardStateHash(Move{}, White)

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("en passant possibility affects hash", func(t *testing.T) {
		board := NewEmptyBoard()
		_ = board.LoadPieces([]Piece{
			NewPiece(Pawn, White, 52),
			NewPiece(Pawn, Black, 53), // adjacent
			NewPiece(King, White, 25),
			NewPiece(King, Black, 95),
		})

		// Move where white pawn didn't double-move
		normalMove := Move{
			Color:  White,
			Symbol: Pawn,
			From:   42,
			To:     52,
		}

		// Move where white pawn double-moved (enables en passant)
		doubleMove := Move{
			Color:  White,
			Symbol: Pawn,
			From:   32,
			To:     52,
		}

		hash1 := board.calculateBoardStateHash(normalMove, Black)
		hash2 := board.calculateBoardStateHash(doubleMove, Black)

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("position repetition detection", func(t *testing.T) {
		board := NewBoard()

		// Initial position hash
		initialHash := board.calculateBoardStateHash(Move{}, White)

		// Make a move and undo it (knight out and back)
		move1 := Move{Color: White, Symbol: Knight, From: 22, To: 42}
		_ = board.ApplyMove(move1)

		move2 := Move{Color: Black, Symbol: Knight, From: 92, To: 72}
		_ = board.ApplyMove(move2)

		move3 := Move{Color: White, Symbol: Knight, From: 42, To: 22}
		_ = board.ApplyMove(move3)

		move4 := Move{Color: Black, Symbol: Knight, From: 72, To: 92}
		_ = board.ApplyMove(move4)

		// After undoing moves, hash should match initial
		finalHash := board.calculateBoardStateHash(Move{}, White)

		assert.Equal(t, initialHash, finalHash)
	})

	t.Run("different piece types at same position produce different hashes", func(t *testing.T) {
		board1 := NewEmptyBoard()
		board2 := NewEmptyBoard()

		// Board 1: Queen at d1
		_ = board1.LoadPieces([]Piece{
			NewPiece(Queen, White, 24, false),
			NewPiece(King, White, 25, false),
			NewPiece(King, Black, 95, false),
		})

		// Board 2: Rook at d1 (different piece, same position)
		_ = board2.LoadPieces([]Piece{
			NewPiece(Rook, White, 24, false),
			NewPiece(King, White, 25, false),
			NewPiece(King, Black, 95, false),
		})

		hash1 := board1.calculateBoardStateHash(Move{}, White)
		hash2 := board2.calculateBoardStateHash(Move{}, White)

		assert.NotEqual(t, hash1, hash2)
	})
}

func TestCalculateCastlingBits(t *testing.T) {
	t.Run("all castling available", func(t *testing.T) {
		board := NewEmptyBoard()
		_ = board.LoadPieces([]Piece{
			// White pieces
			NewPiece(King, White, 25, false),
			NewPiece(Rook, White, 21, false),
			NewPiece(Rook, White, 28, false),
			// Black pieces
			NewPiece(King, Black, 95, false),
			NewPiece(Rook, Black, 91, false),
			NewPiece(Rook, Black, 98, false),
		})

		bits := board.calculateCastlingBits()

		// All castling available: 8 + 4 + 2 + 1 = 15
		assert.Equal(t, uint64(15), bits, "all castling rights should give bits value 15")
	})

	t.Run("no castling available", func(t *testing.T) {
		board := NewEmptyBoard()
		_ = board.LoadPieces([]Piece{
			NewPiece(King, White, 25, true), // Moved
			NewPiece(Rook, White, 21, true), // Moved
			NewPiece(Rook, White, 28, true), // Moved
			NewPiece(King, Black, 95, true), // Moved
			NewPiece(Rook, Black, 91, true), // Moved
			NewPiece(Rook, Black, 98, true), // Moved
		})

		bits := board.calculateCastlingBits()

		assert.Equal(t, uint64(0), bits, "no castling rights should give bits value 0")
	})

	t.Run("only white kingside castling", func(t *testing.T) {
		board := NewEmptyBoard()
		_ = board.LoadPieces([]Piece{
			NewPiece(King, White, 25, false),
			NewPiece(Rook, White, 28, false), // Only kingside rook unmoved
			NewPiece(King, Black, 95, true),
		})

		bits := board.calculateCastlingBits()

		// White kingside: bit 0 = 1
		assert.Equal(t, uint64(1), bits, "white kingside only should give bits value 1")
	})

	t.Run("only black queenside castling", func(t *testing.T) {
		board := NewEmptyBoard()
		_ = board.LoadPieces([]Piece{
			NewPiece(King, White, 25, true),
			NewPiece(King, Black, 95, false),
			NewPiece(Rook, Black, 91, false), // Only black queenside rook unmoved
		})

		bits := board.calculateCastlingBits()

		// Black queenside: bit 3 = 8
		assert.Equal(t, uint64(8), bits, "black queenside only should give bits value 8")
	})
}

func TestEnPassantPossibleAtPosition(t *testing.T) {
	t.Run("no en passant when last move wasn't pawn double-move", func(t *testing.T) {
		board := NewEmptyBoard()
		_ = board.LoadPieces([]Piece{
			NewPiece(Pawn, White, 52, false),
			NewPiece(King, White, 25, false),
			NewPiece(King, Black, 95, false),
		})

		move := Move{
			Color:  White,
			Symbol: Pawn,
			From:   42,
			To:     52, // Single move
		}

		west, east := board.enPassantPossibleAtPosition(move, Black)

		assert.Equal(t, -1, west, "no en passant west")
		assert.Equal(t, -1, east, "no en passant east")
	})

	t.Run("en passant possible when pawn double-moves with adjacent enemy pawn(east)", func(t *testing.T) {
		board := NewEmptyBoard()
		_ = board.LoadPieces([]Piece{
			NewPiece(Pawn, White, 52, false), // Just moved here
			NewPiece(Pawn, Black, 53, false), // Adjacent black pawn (east)
			NewPiece(King, White, 25, false),
			NewPiece(King, Black, 95, false),
		})

		move := Move{
			Color:  White,
			Symbol: Pawn,
			From:   32,
			To:     52, // Double move
		}

		west, east := board.enPassantPossibleAtPosition(move, Black)

		// Black pawn at 53 can capture en passant
		assert.Equal(t, -1, west, "no pawn to the west")
		assert.Equal(t, 53, east, "black pawn at position 53 can capture en passant")
	})

	t.Run("en passant possible when pawn double-moves with adjacent enemy pawn(west)", func(t *testing.T) {
		board := NewEmptyBoard()
		_ = board.LoadPieces([]Piece{
			NewPiece(Pawn, White, 52, false), // Just moved here
			NewPiece(Pawn, Black, 51, false), // Adjacent black pawn (west)
			NewPiece(King, White, 25, false),
			NewPiece(King, Black, 95, false),
		})

		move := Move{
			Color:  White,
			Symbol: Pawn,
			From:   32,
			To:     52, // Double move
		}

		west, east := board.enPassantPossibleAtPosition(move, Black)

		// Black pawn at 53 can capture en passant
		assert.Equal(t, 51, west, "black pawn at position 51 can capture en passant")
		assert.Equal(t, -1, east, "no pawn to the east")
	})

	t.Run("no en passant when active color is same as moving pawn", func(t *testing.T) {
		board := NewEmptyBoard()
		_ = board.LoadPieces([]Piece{
			NewPiece(Pawn, White, 52, false),
			NewPiece(Pawn, Black, 53, false),
			NewPiece(King, White, 25, false),
			NewPiece(King, Black, 95, false),
		})

		move := Move{
			Color:  White,
			Symbol: Pawn,
			From:   32,
			To:     52,
		}

		// Active color is White (same as moving pawn)
		west, east := board.enPassantPossibleAtPosition(move, White)

		assert.Equal(t, -1, west, "no en passant for same color")
		assert.Equal(t, -1, east, "no en passant for same color")
	})
}
