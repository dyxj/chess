package engine

type Game struct {
	board           *Board
	turn            Color
	whiteValidMoves []Move
	blackValidMoves []Move
	whiteGraveyard  []Symbol
	blackGraveyard  []Symbol
	// zeroes if a pawn moves or a capture occurs
	drawCounter int
}

func NewGame() *Game {
	board := NewBoard()

	return &Game{
		board:          board,
		turn:           White,
		whiteGraveyard: make([]Symbol, 0, 16),
		blackGraveyard: make([]Symbol, 0, 16),
		drawCounter:    0,
	}
}
