package game

import "github.com/dyxj/chess/internal/engine"

type Board interface {
	ApplyMove(m engine.Move) error
	UndoLastMove() bool
	Piece(c engine.Color, s engine.Symbol, position int) (engine.Piece, bool)
	Pieces(c engine.Color) []engine.Piece
	ActiveColor() engine.Color
	GeneratePieceLegalMoves(p engine.Piece) ([]engine.Move, error)
	HasLegalMoves(c engine.Color) bool
	IsCheck(c engine.Color) bool
	Grid() string
	GridRaw() [64]int
}
