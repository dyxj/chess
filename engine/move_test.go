package engine

import (
	"fmt"
	"testing"
)

func TestGenerateBasicMoves(t *testing.T) {
	board := NewBoard()
	fmt.Println(board.GridString())
	rook := NewPiece(Rook, White, 41)

	board.cells[41] = boardSymbolPiece(rook)
	fmt.Println(board.GridString())

	moves := generateBasicMoves(board, rook)
	fmt.Println(board.GridString())

	fmt.Println(moves)
}
