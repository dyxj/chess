package game

type State int

const (
	InProgress State = iota + 1
	Checkmate
	Stalemate
	Draw
)

func (s State) String() string {
	switch s {
	case InProgress:
		return "In Progress"
	case Checkmate:
		return "Checkmate"
	case Stalemate:
		return "Stalemate"
	case Draw:
		return "Draw"
	default:
		return "Unknown"
	}
}

func (g *Game) calculateGameState() State {
	activeColor := g.b.ActiveColor()

	if g.b.HasLegalMoves(activeColor) {
		return InProgress
	}

	if g.b.IsCheck(activeColor) {
		g.winner = activeColor.Opposite()
		return Checkmate
	}

	return Stalemate
}
