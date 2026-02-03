package engine

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoard_SaveAndLoad(t *testing.T) {
	originalBoard := NewBoard()
	moves := []Move{
		{Color: White, Symbol: Pawn, From: 34, To: 54},
		{Color: Black, Symbol: Pawn, From: 84, To: 64},
		{Color: White, Symbol: Knight, From: 22, To: 43},
	}

	for _, move := range moves {
		err := originalBoard.ApplyMove(move)
		assert.NoError(t, err)
	}

	var buf bytes.Buffer
	err := originalBoard.Save(&buf)
	assert.NoError(t, err)

	loadedBoard := NewEmptyBoard()

	err = loadedBoard.Load(&buf)
	assert.NoError(t, err)

	assert.Equal(t, originalBoard.cells, loadedBoard.cells)
	assert.Equal(t, originalBoard.whitePieces, loadedBoard.whitePieces)
	assert.Equal(t, originalBoard.blackPieces, loadedBoard.blackPieces)
	assert.Equal(t, originalBoard.whiteKingPos, loadedBoard.whiteKingPos)
	assert.Equal(t, originalBoard.blackKingPos, loadedBoard.blackKingPos)
	assert.Equal(t, originalBoard.roundHistory, loadedBoard.roundHistory)
	assert.Equal(t, originalBoard.activeColor, loadedBoard.activeColor)
	assert.Equal(t, originalBoard.graveyard, loadedBoard.graveyard)
	assert.Equal(t, originalBoard.drawCounter, loadedBoard.drawCounter)
	assert.Equal(t, originalBoard.boardStateHashMapCount, loadedBoard.boardStateHashMapCount)
}
