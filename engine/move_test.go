package engine_test

import (
	. "github.com/dyxj/chess/engine"
)

// generateExpectedMoves generates moves from [from,to] in the given direction.
// takes first capture and assigns it to last move.
func generateExpectedMoves(symbol Symbol, color Color, from int, to int, direction Direction, capture ...Symbol) []Move {
	moves := make([]Move, 0)

	i := from
	for {
		if direction > 0 {
			if i >= to {
				break
			}
		} else {
			if i <= to {
				break
			}
		}
		moves = append(moves, Move{
			Color:  color,
			Symbol: symbol,
			From:   from,
			To:     i + int(direction),
		})
		i += int(direction)
	}

	if len(capture) > 0 && len(moves) > 0 {
		moves[len(moves)-1].Captured = capture[0]
	}
	return moves
}
