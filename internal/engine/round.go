package engine

// round
// Move: applied move
// PrevDrawCounter draw counter from previous round
// BoardStateHash calculated after move applied and with opposite color
type round struct {
	Move            Move
	PrevDrawCounter int
	BoardStateHash  uint64
}
