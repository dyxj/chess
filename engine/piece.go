package engine

type Direction int

const (
	N  Direction = 10
	S  Direction = -10
	E  Direction = 1
	W  Direction = -1
	NE Direction = N + E // 11
	NW Direction = N + W // 9
	SE Direction = S + E // -9
	SW Direction = S + W // -11
)

type Color int

const (
	White Color = 1
	Black Color = -1
)

func (c Color) Opposite(i Color) Color {
	return i * -1
}

type Symbol int

const (
	Pawn   Symbol = 1
	Knight Symbol = 2
	Bishop Symbol = 3
	Rook   Symbol = 4
	Queen  Symbol = 5
	King   Symbol = 6
)

type Piece struct {
	symbol   Symbol
	color    Color
	position int
	hasMoved bool
}

func NewPiece(
	symbol Symbol,
	color Color,
	position int,
) *Piece {
	return &Piece{
		symbol:   symbol,
		color:    color,
		position: position,
		hasMoved: false,
	}
}

var pieceBasicDirections = map[Symbol][]Direction{
	Pawn: {N, NE, NW}, // pawn needs spe
	Knight: {
		N + N + E,
		N + N + W,
		S + S + E,
		S + S + W,
		E + E + N,
		E + E + S,
		W + W + N,
		W + W + S,
	},
	Bishop: {NE, NW, SE, SW},
	Rook:   {N, S, E, W},
	Queen:  {N, S, E, W, NE, NW, SE, SW},
	King:   {N, S, E, W, NE, NW, SE, SW},
}
