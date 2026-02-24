package engine

import "github.com/dyxj/chess/pkg/mathx"

var pawnWhiteDirections = []Direction{N, N + N, NE, NW}
var pawnBlackDirections = []Direction{S, S + S, SE, SW}

func pawnMoveDirections(color Color, hasMoved bool) []Direction {
	moveEndIndex := 2
	if hasMoved {
		moveEndIndex = 1
	}
	if color == White {
		return pawnWhiteDirections[:moveEndIndex]
	}
	return pawnBlackDirections[:moveEndIndex]
}

func pawnCaptureDirections(color Color) []Direction {
	if color == White {
		return pawnWhiteDirections[2:]
	}
	return pawnBlackDirections[2:]
}

func (b *Board) generatePseudoLegalPawnMoves(piece Piece) ([]Move, error) {
	moves := make([]Move, 0, maxMovesByPiece[piece.symbol])

	// Pawn move
	moveDirections := pawnMoveDirections(piece.color, piece.HasMoved())
	for i := 0; i < len(moveDirections); i++ {
		direction := moveDirections[i]
		nextPos := piece.position + int(direction)

		// sentinel should never happen due to promotion
		if !b.IsEmpty(nextPos) || b.IsSentinel(nextPos) {
			// if first forward is not empty, skip the double lastMove forward
			break
		}

		if piece.color == White && nextPos >= 91 && nextPos <= 98 {
			moves = append(moves, b.generatePawnPromotionMoves(piece, nextPos, 0)...)
			continue
		}
		if piece.color == Black && nextPos >= 21 && nextPos <= 28 {
			moves = append(moves, b.generatePawnPromotionMoves(piece, nextPos, 0)...)
			continue
		}

		moves = append(moves, Move{
			Color:    piece.color,
			Symbol:   piece.symbol,
			From:     piece.position,
			To:       nextPos,
			Captured: 0,
		})
	}

	// Pawn capture
	captureDirections := pawnCaptureDirections(piece.color)
	for i := 0; i < len(captureDirections); i++ {
		direction := captureDirections[i]
		nextPos := piece.position + int(direction)

		// sentinel should never happen due to promotion
		if b.IsEmpty(nextPos) || b.Color(nextPos) == piece.color || b.IsSentinel(nextPos) {
			continue
		}

		captured := b.Symbol(nextPos)

		if piece.color == White && nextPos >= 91 && nextPos <= 98 {
			moves = append(moves, b.generatePawnPromotionMoves(piece, nextPos, captured)...)
			continue
		}
		if piece.color == Black && nextPos >= 21 && nextPos <= 28 {
			moves = append(moves, b.generatePawnPromotionMoves(piece, nextPos, captured)...)
			continue
		}

		moves = append(moves, Move{
			Color:    piece.color,
			Symbol:   piece.symbol,
			From:     piece.position,
			To:       nextPos,
			Captured: captured,
		})
	}

	// en passant
	enPassantMove, isEnPassant := b.generateEnPassantMovesIfEligible(piece)
	if isEnPassant {
		moves = append(moves, enPassantMove)
	}

	return moves, nil
}

var promotionSymbols = []Symbol{Queen, Rook, Bishop, Knight}

func (b *Board) generatePawnPromotionMoves(
	piece Piece,
	nextPos int,
	captured Symbol,
) []Move {
	moves := make([]Move, 0, 4)
	for _, promoSymbol := range promotionSymbols {
		moves = append(moves, Move{
			Color:     piece.color,
			Symbol:    piece.symbol,
			From:      piece.position,
			To:        nextPos,
			Captured:  captured,
			Promotion: promoSymbol,
		})
	}
	return moves
}

func (b *Board) generateEnPassantMovesIfEligible(piece Piece) (Move, bool) {
	lastMove, found := b.LastMove()
	if !found {
		return Move{}, false
	}

	if lastMove.Symbol == Pawn &&
		lastMove.Color != piece.color &&
		// double move
		mathx.AbsInt(lastMove.To-lastMove.From) == 20 &&
		// is east or west of current piece
		mathx.AbsInt(piece.position-lastMove.To) == 1 &&
		// check board symbol is pawn and color is opposite
		// board and move history should be sync, but check just in case
		(b.Symbol(lastMove.To) == Pawn && b.Color(lastMove.To) != piece.color) {
		return Move{
			Color:       piece.color,
			Symbol:      piece.symbol,
			From:        piece.position,
			To:          (lastMove.To + lastMove.From) / 2,
			Captured:    Pawn,
			IsEnPassant: true,
		}, true
	}

	return Move{}, false
}
