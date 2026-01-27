package engine

import "github.com/dyxj/chess/pkg/mathx"

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

var pawnWhiteDirections = []Direction{N + N, N, NE, NW}
var pawnBlackDirections = []Direction{S + S, S, SE, SW}

func generatePseudoLegalPawnMoves(board *Board, piece *Piece) ([]Move, error) {
	moves := make([]Move, 0, maxMovesByPiece[piece.symbol])
	var pawnDirections []Direction
	if piece.color == White {
		pawnDirections = pawnWhiteDirections
	} else {
		pawnDirections = pawnBlackDirections
	}

	moveIndex := 0
	if piece.hasMoved {
		// skip double move
		moveIndex = 1
	}

	// Pawn move
	for ; moveIndex < 2; moveIndex++ {
		direction := pawnDirections[moveIndex]
		nextPos := piece.position + int(direction)

		if board.IsSentinel(nextPos) || !board.IsEmpty(nextPos) {
			continue
		}

		if piece.color == White && piece.position >= 91 && piece.position <= 98 {
			appendPawnPromotionMoves(piece, nextPos, 0, moves)
			continue
		}
		if piece.color == Black && piece.position >= 21 && piece.position <= 28 {
			appendPawnPromotionMoves(piece, nextPos, 0, moves)
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

		if board.IsEmpty(nextPos) || board.Color(nextPos) == piece.color || board.IsSentinel(nextPos) {
			continue
		}

		captured := board.Symbol(nextPos)

		if piece.color == White && piece.position >= 91 && piece.position <= 98 {
			appendPawnPromotionMoves(piece, nextPos, captured, moves)
			continue
		}
		if piece.color == Black && piece.position >= 21 && piece.position <= 28 {
			appendPawnPromotionMoves(piece, nextPos, captured, moves)
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
	move, found := board.lastMove()
	if found {
		if move.Symbol == Pawn && mathx.AbsInt(move.To-move.From) == 20 {
			moves = append(moves, Move{
				Color:       piece.color,
				Symbol:      piece.symbol,
				From:        piece.position,
				To:          move.To - int(pawnDirections[1]),
				Captured:    Pawn,
				IsEnPassant: true,
			})
		}
	}

	return moves, nil
}

func appendPawnPromotionMoves(
	piece *Piece,
	nextPos int,
	captured Symbol,
	moves []Move,
) {
	for _, promoSymbol := range []Symbol{Queen, Rook, Bishop, Knight} {
		moves = append(moves, Move{
			Color:     piece.color,
			Symbol:    piece.symbol,
			From:      piece.position,
			To:        nextPos,
			Captured:  captured,
			Promotion: promoSymbol,
		})
	}
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
