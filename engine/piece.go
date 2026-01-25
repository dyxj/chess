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
}

func (p *Piece) Symbol() Symbol {
	return p.symbol
}

func (p *Piece) Color() Color {
	return p.color
}

func (p *Piece) BoardSymbol() int {
	return int(p.symbol) * int(p.color)
}

func (p *Piece) SetPosition(pos int) {
	p.position = pos
}

func (p *Piece) Position() int {
	return p.position
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
	}
}
