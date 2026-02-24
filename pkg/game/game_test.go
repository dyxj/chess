package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGame_fileRankToIndex(t *testing.T) {
	g := &Game{}

	tt := []struct {
		file   byte
		rank   byte
		expect int
	}{
		{file: 'a', rank: '1', expect: 0},
		{file: 'b', rank: '1', expect: 1},
		{file: 'c', rank: '1', expect: 2},
		{file: 'd', rank: '1', expect: 3},
		{file: 'e', rank: '1', expect: 4},
		{file: 'f', rank: '1', expect: 5},
		{file: 'g', rank: '1', expect: 6},
		{file: 'h', rank: '1', expect: 7},

		{file: 'a', rank: '2', expect: 8},
		{file: 'b', rank: '2', expect: 9},
		{file: 'c', rank: '2', expect: 10},
		{file: 'd', rank: '2', expect: 11},
		{file: 'e', rank: '2', expect: 12},
		{file: 'f', rank: '2', expect: 13},
		{file: 'g', rank: '2', expect: 14},
		{file: 'h', rank: '2', expect: 15},

		{file: 'a', rank: '3', expect: 16},
		{file: 'b', rank: '3', expect: 17},
		{file: 'c', rank: '3', expect: 18},
		{file: 'd', rank: '3', expect: 19},
		{file: 'e', rank: '3', expect: 20},
		{file: 'f', rank: '3', expect: 21},
		{file: 'g', rank: '3', expect: 22},
		{file: 'h', rank: '3', expect: 23},

		{file: 'a', rank: '4', expect: 24},
		{file: 'b', rank: '4', expect: 25},
		{file: 'c', rank: '4', expect: 26},
		{file: 'd', rank: '4', expect: 27},
		{file: 'e', rank: '4', expect: 28},
		{file: 'f', rank: '4', expect: 29},
		{file: 'g', rank: '4', expect: 30},
		{file: 'h', rank: '4', expect: 31},

		{file: 'a', rank: '5', expect: 32},
		{file: 'b', rank: '5', expect: 33},
		{file: 'c', rank: '5', expect: 34},
		{file: 'd', rank: '5', expect: 35},
		{file: 'e', rank: '5', expect: 36},
		{file: 'f', rank: '5', expect: 37},
		{file: 'g', rank: '5', expect: 38},
		{file: 'h', rank: '5', expect: 39},

		{file: 'a', rank: '6', expect: 40},
		{file: 'b', rank: '6', expect: 41},
		{file: 'c', rank: '6', expect: 42},
		{file: 'd', rank: '6', expect: 43},
		{file: 'e', rank: '6', expect: 44},
		{file: 'f', rank: '6', expect: 45},
		{file: 'g', rank: '6', expect: 46},
		{file: 'h', rank: '6', expect: 47},

		{file: 'a', rank: '7', expect: 48},
		{file: 'b', rank: '7', expect: 49},
		{file: 'c', rank: '7', expect: 50},
		{file: 'd', rank: '7', expect: 51},
		{file: 'e', rank: '7', expect: 52},
		{file: 'f', rank: '7', expect: 53},
		{file: 'g', rank: '7', expect: 54},
		{file: 'h', rank: '7', expect: 55},

		{file: 'a', rank: '8', expect: 56},
		{file: 'b', rank: '8', expect: 57},
		{file: 'c', rank: '8', expect: 58},
		{file: 'd', rank: '8', expect: 59},
		{file: 'e', rank: '8', expect: 60},
		{file: 'f', rank: '8', expect: 61},
		{file: 'g', rank: '8', expect: 62},
		{file: 'h', rank: '8', expect: 63},
	}

	for _, tc := range tt {
		t.Run(string(tc.file)+string(tc.rank), func(t *testing.T) {
			index := g.fileRankToIndex(tc.file, tc.rank)
			assert.Equal(t, tc.expect, index)
		})
	}
}
