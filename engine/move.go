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

func GeneratePseudoLegalMoves(
	board *Board,
	piece *Piece,
) ([]Move, error) {
	if board.Symbol(piece.position) != piece.symbol {
		return nil, ErrPieceNotFound
	}
	if board.Color(piece.position) != piece.color {
		return nil, ErrPieceNotFound
	}

	if piece.symbol == Pawn {
		return generatePseudoLegalPawnMoves(board, piece)
	}

	return generatePseudoLegalMoves(board, piece)
}

// TODO missing special moves (castling)
func generatePseudoLegalMoves(
	board *Board,
	piece *Piece,
) ([]Move, error) {
	moves := make([]Move, 0, maxMovesByPiece[piece.symbol])

	for _, direction := range pieceBasicDirections[piece.symbol] {
		currentPos := piece.position
		for {
			nextPos := currentPos + int(direction)

			if board.IsSentinel(nextPos) || board.Color(nextPos) == piece.color {
				break
			}

			moves = append(moves,
				Move{
					Color:    piece.color,
					Symbol:   piece.symbol,
					From:     piece.position,
					To:       nextPos,
					Captured: board.Symbol(nextPos), // 0 if sentinel or empty
				})

			if !board.IsEmpty(nextPos) || !isSlidingPiece[piece.symbol] {
				break
			}

			currentPos = nextPos
		}
	}
	return moves, nil
}
