package engine

type Move struct {
	Color  Color
	Symbol Symbol
	From   int
	To     int

	IsCastling bool
	RookFrom   int
	RookTo     int

	Captured    Symbol
	Promotion   Symbol
	IsEnPassant bool
}

func GenerateBasicMoves(
	board *Board,
	piece *Piece,
) []Move {
	var moves []Move

	for _, direction := range pieceBasicDirections[piece.symbol] {
		currentPos := piece.position
		hasNext := true
		for hasNext {
			nextPos := currentPos + int(direction)
			if board.IsSentinel(nextPos) || board.Color(nextPos) == piece.color {
				break
			}

			move := Move{
				Color:       piece.color,
				Symbol:      piece.symbol,
				From:        currentPos,
				To:          nextPos,
				IsCastling:  false,
				RookFrom:    0,
				RookTo:      0,
				Captured:    0,
				Promotion:   0,
				IsEnPassant: false,
			}

			if !board.IsEmpty(nextPos) {
				move.Captured = board.Symbol(nextPos)
				hasNext = false
			}
			if !isSlidingPiece[piece.symbol] {
				hasNext = false
			}

			moves = append(moves, move)
			board.applyMove(move)

			currentPos = nextPos
		}
		board.undoMoves(moves)
	}
	return moves
}
