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

func GenerateLegalMoves(board *Board, color Color) ([]Move, error) {
	moves := make([]Move, 0, maxMovesAllPieces)
	pieces := board.Pieces(color)
	for _, piece := range pieces {
		var err error
		moves, err = GeneratePiecePseudoLegalMoves(board, piece)
		// panic used here as it is a programmer error if board and piece list is out of sync
		panic(err)
	}

	legalCount := 0
	for i, m := range moves {
		board.applyMovePos(m)
		if !board.isKingUnderAttack(color) {
			moves[legalCount] = moves[i]
			legalCount++
		}
		board.undoMovePos(m)
	}

	return moves[:legalCount], nil
}

func GeneratePiecePseudoLegalMoves(
	board *Board,
	piece Piece,
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

	moves := generatePiecePseudoLegalMoves(board, piece)

	moves = append(moves, generateCastlingMoves(board, piece)...)

	return moves, nil
}

func generatePiecePseudoLegalMoves(
	board *Board,
	piece Piece,
) []Move {
	moves := make([]Move, 0, maxMovesByPiece[piece.symbol])

	for _, direction := range pieceDirections[piece.symbol] {
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
	return moves
}

// generateCastlingMoves
// - if king is King and if it hasn't moved
// - if the path between king and rook is clear and rooks haven't moved
// - if king is not checked
// - if king destination is not under attack
// Generates castling moves if all conditions are met
func generateCastlingMoves(board *Board, king Piece) []Move {
	if king.symbol != King || king.hasMoved {
		return nil
	}
	pieces := board.Pieces(king.color)
	if board.isKingUnderAttack(king.color) {
		return nil
	}

	moves := make([]Move, 0, 2)
	rooksFound := 0
	for i := 0; i < len(pieces) && rooksFound < 2; i++ {
		if pieces[i].symbol != Rook {
			continue
		}
		rooksFound++
		if pieces[i].hasMoved {
			continue
		}
		rook := pieces[i]

		direction := E
		if king.position > rook.position {
			direction = W
		}

		pathClear := true
		for nextPos := king.position + int(direction); nextPos != rook.position; nextPos += int(direction) {
			// error in board setup or irrelevant if sentinel is found
			if !board.IsEmpty(nextPos) {
				pathClear = false
				break
			}
		}

		if !pathClear {
			continue
		}

		kingNextPos := king.position + int(direction)*2
		if board.isUnderAttack(kingNextPos, king.color) {
			continue
		}

		moves = append(moves, Move{
			Color:      king.color,
			Symbol:     king.symbol,
			From:       king.position,
			To:         kingNextPos,
			IsCastling: true,
			RookFrom:   rook.position,
			RookTo:     king.position + int(direction),
		})
	}

	return moves
}
