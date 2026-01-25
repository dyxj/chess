package engine

type Color int

const (
	White Color = 1
	Black Color = -1
)

type Symbol int

const (
	Pawn   Symbol = 1
	Knight Symbol = 2
	Bishop Symbol = 3
	Rook   Symbol = 4
	Queen  Symbol = 5
	King   Symbol = 6
)

type Piece interface {
	Symbol() Symbol
	Color() Color
	AllowedMovements() (direction []Direction, canSlide bool)
	Position() int
	BoardSymbol() int
}
