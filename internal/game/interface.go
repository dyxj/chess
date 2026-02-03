package game

import "github.com/dyxj/chess/internal/engine"

type Board interface {
	ApplyMove(m engine.Move) error
	GeneratePieceLegalMoves(p engine.Piece) ([]engine.Move, error)
	UndoLastMove() bool
	Piece(c engine.Color, s engine.Symbol, position int) (engine.Piece, bool)
	ActiveColor() engine.Color
	Grid() string
	GridRaw() [64]int
}
