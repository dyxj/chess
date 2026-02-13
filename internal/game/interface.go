package game

import "github.com/dyxj/chess/internal/engine"

//go:generate go run go.uber.org/mock/mockgen -destination=mock_interface_test.go -package=$GOPACKAGE . Board
type Board interface {
	ApplyMove(m engine.Move) error
	UndoLastMove() bool
	LastMove() (engine.Move, bool)
	Piece(c engine.Color, s engine.Symbol, position int) (engine.Piece, bool)
	Symbol(pos int) engine.Symbol
	ActiveColor() engine.Color
	GeneratePieceLegalMoves(p engine.Piece) ([]engine.Move, error)
	HasLegalMoves(c engine.Color) bool
	IsCheck(c engine.Color) bool
	GridRaw() [64]int
	Is100MoveDraw() bool
	Is3FoldDraw() bool
	MoveCount() int
}
