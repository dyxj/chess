package game

type State int

const (
	StateInProgress State = iota + 1
	StateCheckmate
	StateStalemate
	StateDraw
)

func (s State) String() string {
	switch s {
	case StateInProgress:
		return "In Progress"
	case StateCheckmate:
		return "Checkmate"
	case StateStalemate:
		return "Stalemate"
	case StateDraw:
		return "Draw"
	default:
		return "Unknown"
	}
}

func (g *Game) calculateGameState() State {
	activeColor := g.b.ActiveColor()

	if g.b.HasLegalMoves(activeColor) {
		return StateInProgress
	}

	if g.b.IsCheck(activeColor) {
		g.winner = activeColor.Opposite()
		return StateCheckmate
	}

	return StateStalemate
}
