package game

import (
	"slices"

	"github.com/dyxj/chess/internal/engine"
)

type Game struct {
	b Board
}

func NewGame(b Board) *Game {
	return &Game{b: b}
}

func (g *Game) ApplyMove(m Move) error {
	engineMove, err := g.validateAndConvertMove(m)
	if err != nil {
		return err
	}

	err = g.b.ApplyMove(engineMove)
	if err != nil {
		return err
	}

	return nil
}

func (g *Game) validateAndConvertMove(m Move) (engine.Move, error) {
	piece, ok := g.b.Piece(m.Color, m.Symbol, m.mbFrom())
	if !ok {
		return engine.Move{}, engine.ErrPieceNotFound
	}

	moves, err := g.b.GeneratePieceLegalMoves(piece)
	if err != nil {
		// board and piece out of sync, should panic due to programmer error
		panic(err)
	}

	moveIndex := slices.IndexFunc(moves, func(move engine.Move) bool {
		if move.From == m.mbFrom() && move.To == m.mbTo() {
			return true
		}
		return false
	})
	if moveIndex == -1 {
		return engine.Move{}, ErrIllegalMove
	}

	return moves[moveIndex], nil
}

func (g *Game) UndoLastMove() bool {
	return g.b.UndoLastMove()
}
