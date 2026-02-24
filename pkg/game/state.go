package game

import "fmt"

type State int

const (
	StateInProgress State = iota + 1
	StateCheckmate
	StateStalemate
	StateDraw
	StateWhiteResign
	StateBlackResign
)

const (
	stateInProgressStr  = "in_progress"
	stateCheckmateStr   = "checkmate"
	stateStalemateStr   = "stalemate"
	stateDrawStr        = "draw"
	stateUnknownStr     = "unknown"
	stateWhiteResignStr = "white_resign"
	stateBlackResignStr = "black_resign"
)

func (s State) String() string {
	switch s {
	case StateInProgress:
		return stateInProgressStr
	case StateCheckmate:
		return stateCheckmateStr
	case StateStalemate:
		return stateStalemateStr
	case StateDraw:
		return stateDrawStr
	case StateWhiteResign:
		return stateWhiteResignStr
	case StateBlackResign:
		return stateBlackResignStr
	default:
		return stateUnknownStr
	}
}

func (s State) IsGameOver() bool {
	return s == StateCheckmate || s == StateStalemate ||
		s == StateDraw || s == StateWhiteResign || s == StateBlackResign
}

func (s State) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

//goland:noinspection GoMixedReceiverTypes
func (s *State) UnmarshalText(text []byte) error {
	str := string(text)
	switch str {
	case stateInProgressStr:
		*s = StateInProgress
	case stateCheckmateStr:
		*s = StateCheckmate
	case stateStalemateStr:
		*s = StateStalemate
	case stateDrawStr:
		*s = StateDraw
	case stateWhiteResignStr:
		*s = StateWhiteResign
	case stateBlackResignStr:
		*s = StateBlackResign
	default:
		return fmt.Errorf("unknown state: %s valid state(in_progress,checkmate,stalemate,draw)", str)
	}
	return nil
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
