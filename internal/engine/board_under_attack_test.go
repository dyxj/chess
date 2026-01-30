package engine

import (
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsKingUnderAttackByKing(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		for _, defenderColor := range Colors {
			for _, direction := range pieceDirections[King] {
				board := NewEmptyBoard()
				dKingPos := 54
				dKing := NewPiece(King, defenderColor, dKingPos)

				aKing := NewPiece(King, defenderColor.Opposite(), dKingPos+int(direction))

				err := board.LoadPieces([]Piece{dKing, aKing})
				assert.NoError(t, err)

				isCheck := board.isKingUnderAttack(defenderColor)
				assert.NoError(t, err)
				assert.Equal(t, true, isCheck)
			}
		}
	})

	t.Run("false", func(t *testing.T) {
		for _, defenderColor := range Colors {
			for _, direction := range pieceDirections[King] {
				board := NewEmptyBoard()
				dKingPos := 54
				dKing := NewPiece(King, defenderColor, dKingPos)

				aKing := NewPiece(King, defenderColor.Opposite(), dKingPos+int(direction)*2)

				err := board.LoadPieces([]Piece{dKing, aKing})
				assert.NoError(t, err)

				isCheck := board.isKingUnderAttack(defenderColor)
				assert.NoError(t, err)
				assert.Equal(t, false, isCheck)
			}
		}
	})

}

func TestIsKingUnderAttackBySliders(t *testing.T) {
	tt := []struct {
		name                         string
		symbol                       Symbol
		successfulAttackingDirection []Direction
	}{
		{
			"rook",
			Rook,
			[]Direction{
				N, E, S, W,
			},
		},
		{
			"bishop",
			Bishop,
			[]Direction{
				NE, SW, SE, NW,
			},
		},
		{
			"queen",
			Queen,
			[]Direction{
				N, NE, E, SE, S, SW, W, NW,
			},
		},
	}

	for _, tc := range tt {
		for _, defenderColor := range Colors {
			for _, direction := range directionCircle {
				t.Run(fmt.Sprintf("%v_%v_%v", tc.name, defenderColor, direction), func(t *testing.T) {
					board := NewEmptyBoard()
					dKingPos := 54
					dKing := NewPiece(King, defenderColor, dKingPos)

					attacker := NewPiece(tc.symbol, defenderColor.Opposite(), dKingPos+int(direction)*3)

					err := board.LoadPieces([]Piece{dKing, attacker})
					assert.NoError(t, err)

					isCheck := board.isKingUnderAttack(defenderColor)
					assert.NoError(t, err)
					assert.Equal(t, slices.Contains(tc.successfulAttackingDirection, direction), isCheck)
				})

				t.Run(fmt.Sprintf("%v_%v_%v_blocked", tc.name, defenderColor, direction), func(t *testing.T) {
					board := NewEmptyBoard()
					dKingPos := 54
					dKing := NewPiece(King, defenderColor, dKingPos)

					attacker := NewPiece(tc.symbol, defenderColor.Opposite(), dKingPos+int(direction)*3)
					blocker := NewPiece(tc.symbol, defenderColor, dKingPos+int(direction))

					err := board.LoadPieces([]Piece{dKing, attacker, blocker})
					assert.NoError(t, err)

					isCheck := board.isKingUnderAttack(defenderColor)
					assert.NoError(t, err)
					assert.Equal(t, false, isCheck)
				})
			}
		}
	}
}

func TestIsKingUnderAttackByKnight(t *testing.T) {
	for _, defenderColor := range Colors {
		for _, direction := range []Direction{
			N + N + E,
			N + N + W,
			S + S + E,
			S + S + W,
			E + E + N,
			E + E + S,
			W + W + N,
			W + W + S,
		} {
			t.Run(fmt.Sprintf("%v_%v", defenderColor, direction), func(t *testing.T) {
				board := NewEmptyBoard()
				dKingPos := 54
				dKing := NewPiece(King, defenderColor, dKingPos)

				attacker := NewPiece(Knight, defenderColor.Opposite(), dKingPos+int(direction))

				err := board.LoadPieces([]Piece{dKing, attacker})
				assert.NoError(t, err)

				isCheck := board.isKingUnderAttack(defenderColor)
				assert.NoError(t, err)
				assert.Equal(t, true, isCheck)
			})

			t.Run(fmt.Sprintf("%v_%v_blocked", defenderColor, direction), func(t *testing.T) {
				board := NewEmptyBoard()
				var pieces []Piece

				dKingPos := 54
				pieces = append(pieces, NewPiece(King, defenderColor, dKingPos))

				pieces = append(pieces, NewPiece(Knight, defenderColor.Opposite(), dKingPos+int(direction)))

				for _, blockerDirection := range directionCircle {
					pieces = append(pieces, NewPiece(Pawn, defenderColor, dKingPos+int(blockerDirection)))
				}

				err := board.LoadPieces(pieces)
				assert.NoError(t, err)

				isCheck := board.isKingUnderAttack(defenderColor)
				assert.NoError(t, err)
				assert.Equal(t, true, isCheck)
			})

		}
	}
}

func TestIsKingUnderAttackByPawn(t *testing.T) {
	whiteValidAttackDirections := []Direction{NW, NE}
	blackValidAttackDirections := []Direction{SW, SE}

	for _, defenderColor := range Colors {
		for _, direction := range directionCircle {
			t.Run(fmt.Sprintf("%v_%v", defenderColor, direction), func(t *testing.T) {
				board := NewEmptyBoard()
				var pieces []Piece

				dKingPos := 54
				pieces = append(pieces, NewPiece(King, defenderColor, dKingPos))

				pieces = append(pieces, NewPiece(Pawn, defenderColor.Opposite(), dKingPos+int(direction)))

				err := board.LoadPieces(pieces)
				assert.NoError(t, err)

				isCheck := board.isKingUnderAttack(defenderColor)
				assert.NoError(t, err)

				validAttackDirections := whiteValidAttackDirections
				if defenderColor.Opposite() == Black {
					validAttackDirections = blackValidAttackDirections
				}
				assert.Equal(t, slices.Contains(validAttackDirections, direction), isCheck)
			})
		}
	}
}
