package game

type State int

const (
	InProgress State = iota + 1
	Checkmate
	Stalemate
	Draw
)

func (g *Game) calculateGameState() State {
	activeColor := g.b.ActiveColor()

	if g.b.HasLegalMoves(activeColor) {
		return InProgress
	}

	if g.b.IsCheck(activeColor) {
		return Checkmate
	}

	return Stalemate
}
