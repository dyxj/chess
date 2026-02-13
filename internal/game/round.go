package game

import "github.com/dyxj/chess/internal/engine"

type RoundResult struct {
	Count       int          `json:"count"`
	MoveResult  *MoveResult  `json:"moveResult,omitempty"`
	State       State        `json:"state"`
	Grid        [64]int      `json:"grid"`
	ActiveColor engine.Color `json:"activeColor"`
}
