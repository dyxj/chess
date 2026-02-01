package game

import "github.com/dyxj/chess/internal/engine"

type Game struct {
	b Board
}

func NewGame(b Board) *Game {
	return &Game{b: b}
}

func (g *Game) ApplyMove(m engine.Move) error {
	err := g.b.ApplyMove(m)
	if err != nil {
		return err
	}

	return nil
}

func (g *Game) validateMove(m engine.Move) error {
	piece, ok := g.b.Piece(m.Color, m.Symbol, m.From)
	if !ok {
		return engine.ErrPieceNotFound
	}

	moves, err := g.b.GeneratePieceLegalMoves(piece)
	if err != nil {
		// board and piece out of sync, should panic due to programmer error
		panic(err)
	}

	for _, move := range moves {
		if m == move {
			return nil
		}
	}
	return ErrIllegalMove
}

func (g *Game) UndoLastMove() bool {
	return g.b.UndoLastMove()
}
