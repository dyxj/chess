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

func (c Color) Opposite() Color {
	return c * -1
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

var isSlidingPiece = map[Symbol]bool{
	Pawn:   false,
	Knight: false,
	Bishop: true,
	Rook:   true,
	Queen:  true,
	King:   false,
}

var pieceBasicDirections = map[Symbol][]Direction{
	Pawn: {},
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

var maxMovesByPiece = map[Symbol]int{
	Queen:  56,
	Rook:   28,
	Bishop: 28,
	Knight: 8,
	King:   8,
	Pawn:   4,
}

func GenerateStartPieces(color Color) []*Piece {
	pp := make([]*Piece, 0, 16)

	pawnStart := 31
	pawnEnd := 38
	if color == Black {
		pawnStart = 81
		pawnEnd = 88
	}

	for i := pawnStart; i <= pawnEnd; i++ {
		pp = append(pp, NewPiece(Pawn, color, i))
	}

	powerPieceBase := 20
	if color == Black {
		powerPieceBase = 90
	}

	pp = append(pp, NewPiece(Rook, color, powerPieceBase+1))
	pp = append(pp, NewPiece(Rook, color, powerPieceBase+8))
	pp = append(pp, NewPiece(Knight, color, powerPieceBase+2))
	pp = append(pp, NewPiece(Knight, color, powerPieceBase+7))
	pp = append(pp, NewPiece(Bishop, color, powerPieceBase+3))
	pp = append(pp, NewPiece(Bishop, color, powerPieceBase+6))
	pp = append(pp, NewPiece(Queen, color, powerPieceBase+4))
	pp = append(pp, NewPiece(King, color, powerPieceBase+5))

	return pp
}
