package engine

type Game struct {
	board          *Board
	whitePieces    []*Piece
	blackPieces    []*Piece
	moveHistory    []Move
	whiteGraveyard []Symbol
	blackGraveyard []Symbol
	// zeroes if a pawn moves or a capture occurs
	drawCounter int
}

func NewGame() *Game {
	board := NewBoard()

	whitePieces, blackPieces := extractPiecesFromBoard(board)

	return &Game{
		board:          board,
		whitePieces:    whitePieces,
		blackPieces:    blackPieces,
		moveHistory:    make([]Move, 0, 256),
		whiteGraveyard: make([]Symbol, 0, 16),
		blackGraveyard: make([]Symbol, 0, 16),
		drawCounter:    0,
	}
}

func extractPiecesFromBoard(board *Board) (whitePieces []*Piece, blackPieces []*Piece) {
	whitePieces = make([]*Piece, 0, 16)
	blackPieces = make([]*Piece, 0, 16)

	for i := 0; i < boardSize; i++ {
		if board.IsEmpty(i) || board.IsSentinel(i) {
			continue
		}
		if board.Color(i) == White {
			whitePieces = append(whitePieces, NewPiece(board.Symbol(i), White, i))
		} else {
			blackPieces = append(blackPieces, NewPiece(board.Symbol(i), Black, i))
		}
	}

	return whitePieces, blackPieces
}
