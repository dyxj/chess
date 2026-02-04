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

func (m Move) hasCaptured() bool {
	return m.Captured != 0
}

func (m Move) hasPromotion() bool {
	return m.Promotion != 0
}

func (m Move) calculateEnPassantCapturedPos() int {
	pawnDirection := pawnMoveDirections(m.Color, true)[0]
	return m.To - int(pawnDirection)
}

func (b *Board) GenerateLegalMoves(color Color) ([]Move, error) {
	moves := make([]Move, 0, maxMovesAllPieces)
	pieces := b.Pieces(color)
	for _, piece := range pieces {
		var err error
		pieceMoves, err := b.GeneratePiecePseudoLegalMoves(piece)
		if err != nil {
			// panic used here as it is a programmer error if b and piece list is out of sync
			panic(err)
		}
		moves = append(moves, pieceMoves...)
	}

	return b.filterLegalMoves(moves, color), nil
}

func (b *Board) GeneratePieceLegalMoves(piece Piece) ([]Move, error) {
	moves, err := b.GeneratePiecePseudoLegalMoves(piece)
	if err != nil {
		return nil, err
	}

	return b.filterLegalMoves(moves, piece.color), nil
}

func (b *Board) filterLegalMoves(moves []Move, color Color) []Move {
	legalCount := 0
	for i, m := range moves {
		b.applyMovePos(m)
		if !b.IsCheck(color) {
			moves[legalCount] = moves[i]
			legalCount++
		}
		b.undoMovePos(m)
	}
	return moves[:legalCount]
}

func (b *Board) GeneratePiecePseudoLegalMoves(piece Piece) ([]Move, error) {
	if b.Symbol(piece.position) != piece.symbol {
		return nil, ErrPieceNotFound
	}
	if b.Color(piece.position) != piece.color {
		return nil, ErrPieceNotFound
	}

	if piece.symbol == Pawn {
		return b.generatePseudoLegalPawnMoves(piece)
	}

	moves := b.generatePiecePseudoLegalMoves(piece)

	moves = append(moves, b.generateCastlingMoves(piece)...)

	return moves, nil
}

func (b *Board) generatePiecePseudoLegalMoves(piece Piece) []Move {
	moves := make([]Move, 0, maxMovesByPiece[piece.symbol])

	for _, direction := range pieceDirections[piece.symbol] {
		currentPos := piece.position
		for {
			nextPos := currentPos + int(direction)

			if b.IsSentinel(nextPos) || b.Color(nextPos) == piece.color {
				break
			}

			moves = append(moves,
				Move{
					Color:    piece.color,
					Symbol:   piece.symbol,
					From:     piece.position,
					To:       nextPos,
					Captured: b.Symbol(nextPos), // 0 if sentinel or empty
				})

			if !b.IsEmpty(nextPos) || !isSlidingPiece[piece.symbol] {
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
func (b *Board) generateCastlingMoves(king Piece) []Move {
	if king.symbol != King || king.HasMoved() {
		return nil
	}
	pieces := b.Pieces(king.color)
	if b.IsCheck(king.color) {
		return nil
	}

	moves := make([]Move, 0, 2)
	rooksFound := 0
	for i := 0; i < len(pieces) && rooksFound < 2; i++ {
		if pieces[i].symbol != Rook {
			continue
		}
		rooksFound++
		if pieces[i].HasMoved() {
			continue
		}
		rook := pieces[i]

		direction := E
		if king.position > rook.position {
			direction = W
		}

		pathClear := true
		for nextPos := king.position + int(direction); nextPos != rook.position; nextPos += int(direction) {
			// error in b setup or irrelevant if sentinel is found
			if !b.IsEmpty(nextPos) {
				pathClear = false
				break
			}
		}

		if !pathClear {
			continue
		}

		kingNextPos := king.position + int(direction)*2
		if b.isUnderAttack(kingNextPos, king.color) {
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
