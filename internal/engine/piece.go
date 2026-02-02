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

var Colors = []Color{White, Black}

func (c Color) Opposite() Color {
	return c * -1
}

func (c Color) String() string {
	switch c {
	case White:
		return "white"
	case Black:
		return "black"
	default:
		return "unknown"
	}
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

var Symbols = []Symbol{Pawn, Knight, Bishop, Rook, Queen, King}

type Piece struct {
	symbol    Symbol
	color     Color
	position  int
	moveCount int
}

func NewPiece(
	symbol Symbol,
	color Color,
	position int,
	hasMoved ...bool,
) Piece {
	moveCount := 0
	if len(hasMoved) > 0 {
		moveCount = 1
	}
	return Piece{
		symbol:    symbol,
		color:     color,
		position:  position,
		moveCount: moveCount,
	}
}

func (p Piece) Symbol() Symbol {
	return p.symbol
}

func (p Piece) Color() Color {
	return p.color
}

func (p Piece) Position() int {
	return p.position
}

func (p Piece) HasMoved() bool {
	return p.moveCount > 0
}

var isSlidingPiece = map[Symbol]bool{
	Pawn:   false,
	Knight: false,
	Bishop: true,
	Rook:   true,
	Queen:  true,
	King:   false,
}

var pieceDirections = map[Symbol][]Direction{
	Pawn: {}, // empty as it has specialized handling
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

var maxMovesAllPieces = maxMovesByPiece[Pawn] +
	maxMovesByPiece[Knight] +
	maxMovesByPiece[Bishop] +
	maxMovesByPiece[Rook] +
	maxMovesByPiece[Queen] +
	maxMovesByPiece[King]

var directionCircle = [8]Direction{
	N,
	NE,
	E,
	SE,
	S,
	SW,
	W,
	NW,
}
var slidingMoversByDirectionCircleIndex = [8][]Symbol{
	{Queen, Rook},
	{Queen, Bishop},
	{Queen, Rook},
	{Queen, Bishop},
	{Queen, Rook},
	{Queen, Bishop},
	{Queen, Rook},
	{Queen, Bishop},
}

func GenerateStartPieces(color Color) []Piece {
	pp := make([]Piece, 0, 16)

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
