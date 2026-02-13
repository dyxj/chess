package game

type RoundResult struct {
	Count      int        `json:"count"`
	MoveResult MoveResult `json:"moveResult"`
	State      State      `json:"state"`
}
