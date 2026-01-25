package engine

import (
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBoard(t *testing.T) {
	b := NewBoard()

	expectedCells := [120]int{
		7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
		7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
		7, 4, 2, 3, 5, 6, 3, 2, 4, 7,
		7, 1, 1, 1, 1, 1, 1, 1, 1, 7,
		7, 0, 0, 0, 0, 0, 0, 0, 0, 7,
		7, 0, 0, 0, 0, 0, 0, 0, 0, 7,
		7, 0, 0, 0, 0, 0, 0, 0, 0, 7,
		7, 0, 0, 0, 0, 0, 0, 0, 0, 7,
		7, -1, -1, -1, -1, -1, -1, -1, -1, 7,
		7, -4, -2, -3, -5, -6, -3, -2, -4, 7,
		7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
		7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
	}

	assert.Equal(t, 120, len(b.cells))
	assert.Equal(t, expectedCells, b.cells)
}

func TestBoard_IsEmpty(t *testing.T) {
	board := genTestBoard()

	assert.True(t, board.IsEmpty(0))
	assert.False(t, board.IsEmpty(1))
	assert.False(t, board.IsEmpty(2))
	assert.False(t, board.IsEmpty(3))
}

func TestBoard_IsSentinel(t *testing.T) {
	board := genTestBoard()

	assert.False(t, board.IsSentinel(0))
	assert.True(t, board.IsSentinel(1))
	assert.False(t, board.IsSentinel(2))
	assert.False(t, board.IsSentinel(3))
}

func genTestBoard() *Board {
	return &Board{
		cells: [120]int{
			0,
			7,
			rand.IntN(5) + 1,
			(rand.IntN(5) + 1) * -1,
		},
	}
}
