package engine

type Game struct {
	board *Board
}

func NewGame() *Game {
	board := NewBoard()

	return &Game{
		board: board,
	}
}
