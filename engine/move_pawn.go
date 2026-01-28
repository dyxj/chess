package engine

import "github.com/dyxj/chess/pkg/mathx"

var pawnWhiteDirections = []Direction{N, N + N, NE, NW}
var pawnBlackDirections = []Direction{S, S + S, SE, SW}

func generatePseudoLegalPawnMoves(board *Board, piece *Piece) ([]Move, error) {
	moves := make([]Move, 0, maxMovesByPiece[piece.symbol])
	var pawnDirections []Direction
	if piece.color == White {
		pawnDirections = pawnWhiteDirections
	} else {
		pawnDirections = pawnBlackDirections
	}

	moveEndIndex := 2
	if piece.hasMoved {
		// skip double lastMove
		moveEndIndex = 1
	}

	// Pawn lastMove
	for moveIndex := 0; moveIndex < moveEndIndex; moveIndex++ {
		direction := pawnDirections[moveIndex]
		nextPos := piece.position + int(direction)

		// sentinel should never happen due to promotion
		if !board.IsEmpty(nextPos) || board.IsSentinel(nextPos) {
			// if first forward is not empty, skip the double lastMove forward
			break
		}

		if piece.color == White && nextPos >= 91 && nextPos <= 98 {
			moves = append(moves, generatePawnPromotionMoves(piece, nextPos, 0)...)
			continue
		}
		if piece.color == Black && nextPos >= 21 && nextPos <= 28 {
			moves = append(moves, generatePawnPromotionMoves(piece, nextPos, 0)...)
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
	for captureIndex := 2; captureIndex < 4; captureIndex++ {
		direction := pawnDirections[captureIndex]
		nextPos := piece.position + int(direction)

		// sentinel should never happen due to promotion
		if board.IsEmpty(nextPos) || board.Color(nextPos) == piece.color || board.IsSentinel(nextPos) {
			continue
		}

		captured := board.Symbol(nextPos)

		if piece.color == White && nextPos >= 91 && nextPos <= 98 {
			moves = append(moves, generatePawnPromotionMoves(piece, nextPos, captured)...)
			continue
		}
		if piece.color == Black && nextPos >= 21 && nextPos <= 28 {
			moves = append(moves, generatePawnPromotionMoves(piece, nextPos, captured)...)
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
	enPassantMove, isEnPassant := generateEnPassantMovesIfEligible(board, piece)
	if isEnPassant {
		moves = append(moves, enPassantMove)
	}

	return moves, nil
}

var promotionSymbols = []Symbol{Queen, Rook, Bishop, Knight}

func generatePawnPromotionMoves(
	piece *Piece,
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

func generateEnPassantMovesIfEligible(
	board *Board,
	piece *Piece,
) (Move, bool) {
	lastMove, found := board.lastMove()
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
		(board.Symbol(lastMove.To) == Pawn && board.Color(lastMove.To) != piece.color) {
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
