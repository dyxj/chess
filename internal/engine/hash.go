package engine

import (
	"hash/fnv"

	"github.com/dyxj/chess/pkg/mathx"
)

// calculateBoardStateHash
func (b *Board) calculateBoardStateHash(move Move, active Color) uint64 {
	h := fnv.New64a()
	hashBytes := make([]byte, len(b.cells)+1)
	// board state
	for i := 0; i < len(b.cells); i++ {
		// =ve values wraps around, but it shouldn't be a problem as it is consistent
		// and clashes are not produced
		hashBytes[i] = byte(i)
	}

	epWest, epEast := b.enPassantPossibleAtPosition(move, active)
	hashBytes = append(hashBytes,
		byte(b.activeColor),
		byte(b.calculateCastlingBits()),
		byte(epWest),
		byte(epEast),
	)

	// implementation doesn't return any error
	_, _ = h.Write(hashBytes)

	return h.Sum64()
}

// calculateCastlingBits given 0000 as representation of castling ability
// bit
// 0: white east
// 1: white west
// 2: black east
// 3: black west
func (b *Board) calculateCastlingBits() uint64 {
	moves := make([]Move, 0, 4)
	for _, color := range Colors {
		pp := b.Pieces(color)
		var king Piece
		for i := 0; i < len(pp); i++ {
			if pp[i].symbol == King {
				king = pp[i]
				break
			}
		}

		moves = append(moves, generateCastlingMoves(b, king)...)
	}

	bits := uint64(0)
	for _, move := range moves {
		switch move.RookFrom {
		case 91:
			bits += 8
		case 98:
			bits += 4
		case 21:
			bits += 2
		case 28:
			bits += 1
		}
	}

	return bits
}

func (b *Board) enPassantPossibleAtPosition(lastMove Move, activeColor Color) (west int, east int) {
	if lastMove.Color == activeColor {
		return -1, -1
	}

	activeColorPawnValue := 1
	if activeColor == Black {
		activeColorPawnValue = -1
	}

	if lastMove.Symbol == Pawn &&
		// double move
		mathx.AbsInt(lastMove.To-lastMove.From) == 20 {
		if b.Value(lastMove.To+1) == activeColorPawnValue {
			east = lastMove.To + 1
		}
		if b.Value(lastMove.To-1) == activeColorPawnValue {
			west = lastMove.To - 1
		}
	}

	return -1, -1
}
