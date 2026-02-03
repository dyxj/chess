package engine

import (
	"fmt"
	"slices"
	"strings"
)

// Board dimensions
const boardWidth = 10  // column
const boardHeight = 12 // row
const boardSize = boardWidth * boardHeight

// Board representation
const (
	EmptyCell    = 0
	SentinelCell = 7
)

// look-up table for index to mailbox
var indexToMailbox = [64]int{
	21, 22, 23, 24, 25, 26, 27, 28,
	31, 32, 33, 34, 35, 36, 37, 38,
	41, 42, 43, 44, 45, 46, 47, 48,
	51, 52, 53, 54, 55, 56, 57, 58,
	61, 62, 63, 64, 65, 66, 67, 68,
	71, 72, 73, 74, 75, 76, 77, 78,
	81, 82, 83, 84, 85, 86, 87, 88,
	91, 92, 93, 94, 95, 96, 97, 98,
}

func IndexToMailbox(cell8x8 int) int {
	return indexToMailbox[cell8x8]
}

/*
Board
|110. 7|111. 7|112. 7|113. 7|114. 7|115. 7|116. 7|117. 7|118. 7|119. 7|
|100. 7|101. 7|102. 7|103. 7|104. 7|105. 7|106. 7|107. 7|108. 7|109. 7|
| 90. 7| 91.-4| 92.-2| 93.-3| 94.-5| 95.-6| 96.-3| 97.-2| 98.-4| 99. 7|
| 80. 7| 81.-1| 82.-1| 83.-1| 84.-1| 85.-1| 86.-1| 87.-1| 88.-1| 89. 7|
| 70. 7| 71. 0| 72. 0| 73. 0| 74. 0| 75. 0| 76. 0| 77. 0| 78. 0| 79. 7|
| 60. 7| 61. 0| 62. 0| 63. 0| 64. 0| 65. 0| 66. 0| 67. 0| 68. 0| 69. 7|
| 50. 7| 51. 0| 52. 0| 53. 0| 54. 0| 55. 0| 56. 0| 57. 0| 58. 0| 59. 7|
| 40. 7| 41. 0| 42. 0| 43. 0| 44. 0| 45. 0| 46. 0| 47. 0| 48. 0| 49. 7|
| 30. 7| 31. 1| 32. 1| 33. 1| 34. 1| 35. 1| 36. 1| 37. 1| 38. 1| 39. 7|
| 20. 7| 21. 4| 22. 2| 23. 3| 24. 5| 25. 6| 26. 3| 27. 2| 28. 4| 29. 7|
| 10. 7| 11. 7| 12. 7| 13. 7| 14. 7| 15. 7| 16. 7| 17. 7| 18. 7| 19. 7|
|  0. 7|  1. 7|  2. 7|  3. 7|  4. 7|  5. 7|  6. 7|  7. 7|  8. 7|  9. 7|
*/
type Board struct {
	cells                  [boardSize]int
	whitePieces            []Piece
	blackPieces            []Piece
	whiteKingPos           int
	blackKingPos           int
	roundHistory           []round
	activeColor            Color
	graveyard              []Piece
	drawCounter            int
	boardStateHashMapCount map[uint64]int
}

// NewBoard creates a new chess board with the initial pieces.
func NewBoard() *Board {
	b := NewEmptyBoard()

	wp := GenerateStartPieces(White)
	err := b.LoadPieces(wp)
	if err != nil {
		panic(err)
	}

	bp := GenerateStartPieces(Black)
	err = b.LoadPieces(bp)
	if err != nil {
		panic(err)
	}

	return b
}

func NewEmptyBoard(activeColor ...Color) *Board {
	cells := [boardSize]int{}
	for pos := 0; pos < boardSize; pos++ {
		cells[pos] = calculateBlankBoardValue(pos)
	}
	ac := White
	if len(activeColor) > 0 {
		ac = activeColor[0]
	}
	b := &Board{
		cells:                  cells,
		whitePieces:            make([]Piece, 0, 16),
		blackPieces:            make([]Piece, 0, 16),
		roundHistory:           make([]round, 0, 256),
		graveyard:              make([]Piece, 0, 32),
		activeColor:            ac,
		drawCounter:            0,
		boardStateHashMapCount: make(map[uint64]int, 256),
	}
	return b
}

func (b *Board) LoadPieces(pp []Piece) error {
	for _, p := range pp {
		err := b.loadPiece(p)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Board) loadPiece(p Piece) error {
	if p.position > len(b.cells)-1 {
		return ErrOutOfBoard
	}
	if b.IsSentinel(p.position) {
		return ErrOutOfBoard
	}
	if !b.IsEmpty(p.position) {
		return ErrOccupied
	}
	b.cells[p.position] = boardSymbolPiece(p)
	if p.symbol == King {
		b.setKingPosition(p.color, p.position)
	}
	if p.color == White {
		b.whitePieces = append(b.whitePieces, p)
	} else {
		b.blackPieces = append(b.blackPieces, p)
	}
	return nil
}

func (b *Board) Pieces(color Color) []Piece {
	if color == White {
		return b.whitePieces
	}
	return b.blackPieces
}

func (b *Board) Piece(color Color, symbol Symbol, pos int) (Piece, bool) {
	for _, p := range b.Pieces(color) {
		if p.symbol == symbol && p.position == pos {
			return p, true
		}
	}
	return Piece{}, false
}

func (b *Board) setPieces(color Color, pp []Piece) {
	if color == White {
		b.whitePieces = pp
	} else {
		b.blackPieces = pp
	}
}

func (b *Board) kingPosition(color Color) int {
	if color == White {
		return b.whiteKingPos
	}
	return b.blackKingPos
}

func (b *Board) setKingPosition(color Color, pos int) {
	if color == White {
		b.whiteKingPos = pos
	} else {
		b.blackKingPos = pos
	}
}

func (b *Board) IsEmpty(pos int) bool {
	return b.cells[pos] == EmptyCell
}
func (b *Board) IsSentinel(pos int) bool {
	return b.cells[pos] == SentinelCell
}

func (b *Board) Color(pos int) Color {
	cv := b.cells[pos]
	if cv == 0 || cv == 7 {
		// risky silent failure
		// options
		// 1. could return ok check
		// 2. log error
		// 3. panic
		// Choosing to ignore relying on right implementation
		return 0
	}
	if cv > 0 {
		return White
	}
	return Black
}

func (b *Board) Symbol(pos int) Symbol {
	cv := b.cells[pos]
	if cv == 0 || cv == 7 {
		// risky silent failure
		// options
		// 1. could return ok check
		// 2. log error
		// 3. panic
		// Choosing to ignore relying on right implementation
		return 0
	}
	if cv > 0 {
		return Symbol(cv)
	}
	return Symbol(-cv)
}

func (b *Board) isKingUnderAttack(color Color) bool {
	kPos := b.kingPosition(color)
	// no king on board, for playground boards
	if kPos == 0 {
		return false
	}
	return b.isUnderAttack(b.kingPosition(color), color)
}

// isUnderAttack check if position is under attack by opponent pieces
// using a backward approach.
func (b *Board) isUnderAttack(pos int, defender Color) bool {
	attacker := defender.Opposite()

	// Attacked by sliders Queen, Rook, Bishop
	for i, direction := range directionCircle {
		if b.isUnderAttackBySlider(pos, direction, attacker, slidingMoversByDirectionCircleIndex[i]) {
			return true
		}
	}

	// Attacked by horse
	for _, direction := range pieceDirections[Knight] {
		if b.isUnderAttackByFixDirection(pos, direction, attacker, Knight) {
			return true
		}
	}

	// Attacked by King
	for _, direction := range pieceDirections[King] {
		if b.isUnderAttackByFixDirection(pos, direction, attacker, King) {
			return true
		}
	}

	// Attacked by Pawn
	for _, direction := range pawnCaptureDirections(attacker) {
		if b.isUnderAttackByFixDirection(pos, direction, attacker, Pawn) {
			return true
		}
	}

	return false
}

func (b *Board) isUnderAttackBySlider(pos int, direction Direction, attacker Color, symbols []Symbol) bool {
	dInt := int(direction)
	attackerPos := pos + dInt
	defender := attacker.Opposite()
	for ; ; attackerPos += dInt {
		if b.IsSentinel(attackerPos) || b.Color(attackerPos) == defender {
			break
		}

		if b.IsEmpty(attackerPos) {
			continue
		}

		if slices.Contains(symbols, b.Symbol(attackerPos)) {
			return true
		}
	}
	return false
}

func (b *Board) isUnderAttackByFixDirection(pos int, direction Direction, attacker Color, symbol Symbol) bool {
	attackerPos := pos + int(direction)
	if b.Symbol(attackerPos) == symbol && b.Color(attackerPos) == attacker {
		return true
	}
	return false
}

func (b *Board) Value(pos int) int {
	return b.cells[pos]
}

func (b *Board) GridFull() string {
	sb := strings.Builder{}
	for x := boardHeight - 1; x >= 0; x-- {
		sb.WriteString("|")
		for y := 0; y < boardWidth; y++ {
			i := x*boardWidth + y

			sb.WriteString(fmt.Sprintf("%2d|", b.Value(i)))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func (b *Board) Grid() string {
	sb := strings.Builder{}
	sb.Grow(208)
	sb.WriteString("|")
	for i := len(indexToMailbox) - 1; i >= 0; i-- {
		v := indexToMailbox[i]
		sb.WriteString(fmt.Sprintf("%2d|", b.Value(v)))
		if v%10 == 1 && v != 21 {
			sb.WriteString("\n|")
		}
	}
	sb.WriteString("\n")
	return sb.String()
}

func (b *Board) GridRaw() [64]int {
	g := [64]int{}
	for i, v := range indexToMailbox {
		g[i] = b.Value(v)
	}
	return g
}

func (b *Board) ApplyMove(m Move) error {
	err := b.validateMoveToApply(m)
	if err != nil {
		return err
	}

	r := round{
		Move:            m,
		PrevDrawCounter: b.drawCounter,
	}

	b.applyMovePos(m)

	b.applyMoveToPieceList(m)

	b.setDrawCounter(m)

	b.activeColor = b.activeColor.Opposite()

	r.BoardStateHash = b.calculateBoardStateHash(m, b.activeColor)
	b.boardStateHashMapCount[r.BoardStateHash]++

	b.addRoundToHistory(r)
	return nil
}

func (b *Board) validateMoveToApply(m Move) error {
	if m.Color != b.activeColor {
		return ErrNotActiveColor
	}
	if m.From > len(b.cells)-1 || m.To > len(b.cells)-1 {
		return ErrOutOfBoard
	}
	if b.IsSentinel(m.From) || b.IsSentinel(m.To) {
		return ErrOutOfBoard
	}
	if b.IsEmpty(m.From) {
		return ErrPieceNotFound
	}
	if b.Color(m.From) != m.Color {
		return ErrPieceNotFound
	}
	if b.Symbol(m.From) != m.Symbol {
		return ErrPieceNotFound
	}
	return nil
}

func (b *Board) UndoLastMove() (ok bool) {
	r, found := b.lastRound()
	if !found {
		return false
	}

	b.undoMovePos(r.Move)

	b.undoMovePieceList(r.Move)

	b.drawCounter = r.PrevDrawCounter

	b.activeColor = b.activeColor.Opposite()

	b.boardStateHashMapCount[r.BoardStateHash]--

	b.removeLastRoundFromHistory()

	return true
}

func (b *Board) applyMovePos(m Move) {
	if m.IsEnPassant {
		b.applyEnPassantMovePos(m)
		return
	}

	b.cells[m.From] = EmptyCell
	if m.hasPromotion() {
		b.cells[m.To] = boardSymbol(m.Promotion, m.Color)
		return
	}

	b.cells[m.To] = boardSymbolMove(m)
	if m.IsCastling {
		b.cells[m.RookFrom] = EmptyCell
		b.cells[m.RookTo] = boardSymbol(Rook, m.Color)
	}

	if m.Symbol == King {
		b.setKingPosition(m.Color, m.To)
	}
}

func (b *Board) undoMovePos(m Move) {
	if m.IsEnPassant {
		b.undoEnPassantMovePos(m)
		return
	}

	b.cells[m.From] = boardSymbolMove(m)
	if m.hasCaptured() {
		b.cells[m.To] = boardSymbol(m.Captured, m.Color.Opposite())
	} else {
		b.cells[m.To] = EmptyCell
	}

	if m.IsCastling {
		b.cells[m.RookFrom] = boardSymbol(Rook, m.Color)
		b.cells[m.RookTo] = EmptyCell
	}

	if m.Symbol == King {
		b.setKingPosition(m.Color, m.From)
	}
}

func (b *Board) applyEnPassantMovePos(m Move) {
	b.cells[m.From] = EmptyCell
	b.cells[m.To] = boardSymbolMove(m)

	b.cells[m.calculateEnPassantCapturedPos()] = EmptyCell
}

func (b *Board) undoEnPassantMovePos(m Move) {
	b.cells[m.From] = boardSymbolMove(m)
	b.cells[m.To] = EmptyCell
	b.cells[m.calculateEnPassantCapturedPos()] = boardSymbol(Pawn, m.Color.Opposite())
}

func (b *Board) applyMoveToPieceList(m Move) {

	pp := b.Pieces(m.Color)

	if m.hasCaptured() {
		xColor := m.Color.Opposite()
		xpp := b.Pieces(xColor)

		capturedPos := m.To
		if m.IsEnPassant {
			capturedPos = m.calculateEnPassantCapturedPos()
		}

		// Remove captured piece
		for i := 0; i < len(xpp); i++ {
			if xpp[i].symbol == m.Captured && xpp[i].position == capturedPos {
				b.graveyard = append(b.graveyard, xpp[i])
				xpp = slices.Delete(xpp, i, i+1)
				break
			}
		}

		b.setPieces(xColor, xpp)
	}

	if m.hasPromotion() {
		// replace pawn with promoted piece
		for i := 0; i < len(pp); i++ {
			if pp[i].symbol == m.Symbol && pp[i].position == m.From {
				pp[i] = NewPiece(m.Promotion, m.Color, m.To, true)
				break
			}
		}
		b.setPieces(m.Color, pp)
		return
	}

	// Normal move: update piece position
	for i := 0; i < len(pp); i++ {
		if pp[i].symbol == m.Symbol && pp[i].position == m.From {
			pp[i].position = m.To
			pp[i].moveCount++
			break
		}
	}

	// update rook position
	if m.IsCastling {
		for i := 0; i < len(pp); i++ {
			if pp[i].symbol == Rook && pp[i].position == m.RookFrom {
				pp[i].position = m.RookTo
				pp[i].moveCount++
				break
			}
		}
	}

	b.setPieces(m.Color, pp)
}

// undoMovePieceList update piece list by reverting move
// panics if last piece in graveyard does not match captured move
func (b *Board) undoMovePieceList(m Move) {
	pp := b.Pieces(m.Color)

	if m.hasCaptured() {
		capturedPiece, hasCaptured := b.popGraveyard()
		if hasCaptured {
			if capturedPiece.symbol != m.Captured {
				// board out of sync, programmer error
				panic("last graveyard symbol does not match move symbol")
			}
			xpp := b.Pieces(capturedPiece.Color())
			xpp = append(xpp, capturedPiece)

			b.setPieces(capturedPiece.Color(), xpp)
		} else {
			panic("expect piece in graveyard but it is empty")
		}
	}

	if m.hasPromotion() {
		// replaced promoted with pawn
		for i := 0; i < len(pp); i++ {
			if pp[i].symbol == m.Promotion && pp[i].position == m.To {
				pp[i] = NewPiece(m.Symbol, m.Color, m.From, true)
				break
			}
		}
		b.setPieces(m.Color, pp)
		return
	}

	// Normal move: revert piece position
	for i := 0; i < len(pp); i++ {
		if pp[i].symbol == m.Symbol && pp[i].position == m.To {
			pp[i].position = m.From
			pp[i].moveCount--
			break
		}
	}

	if m.IsCastling {
		for i := 0; i < len(pp); i++ {
			if pp[i].symbol == Rook && pp[i].position == m.RookTo {
				pp[i].position = m.RookFrom
				pp[i].moveCount--
				break
			}
		}
	}

	b.setPieces(m.Color, pp)
}

func (b *Board) addRoundToHistory(r round) {
	b.roundHistory = append(b.roundHistory, r)
}

func (b *Board) removeLastRoundFromHistory() {
	// reslice excluding last
	b.roundHistory = b.roundHistory[:len(b.roundHistory)-1]
}

func (b *Board) lastRound() (round, bool) {
	if len(b.roundHistory) == 0 {
		return round{}, false
	}
	return b.roundHistory[len(b.roundHistory)-1], true
}

func (b *Board) lastMove() (move Move, found bool) {
	round, found := b.lastRound()
	if !found {
		return Move{}, false
	}
	return round.Move, true
}

func (b *Board) Is3FoldDraw(hash uint64) bool {
	count := b.boardStateHashMapCount[hash]
	if count >= 3 {
		return true
	}
	return false
}

// setDrawCounter
// Increments draw counter by 1 if move is not a pawn move and not a capture
func (b *Board) setDrawCounter(m Move) {
	if m.Captured != 0 || m.Symbol == Pawn {
		b.drawCounter = 0
		return
	}
	b.drawCounter++
}

func (b *Board) Is100MoveDraw() bool {
	return b.drawCounter >= 100
}

func (b *Board) popGraveyard() (Piece, bool) {
	if len(b.graveyard) == 0 {
		return Piece{}, false
	}

	lastPiece := b.graveyard[len(b.graveyard)-1]

	// reslice excluding last
	b.graveyard = b.graveyard[:len(b.graveyard)-1]

	return lastPiece, true
}

func (b *Board) ActiveColor() Color {
	return b.activeColor
}

func boardSymbolPiece(p Piece) int {
	return int(p.symbol) * int(p.color)
}

func boardSymbolMove(m Move) int {
	return int(m.Symbol) * int(m.Color)
}

func boardSymbol(s Symbol, color Color) int {
	return int(s) * int(color)
}

func calculateBlankBoardValue(pos int) int {
	if (pos >= 0 && pos <= 19) ||
		(pos%10 == 0) ||
		(pos%10 == 9) ||
		(pos >= 100 && pos <= 119) {
		return SentinelCell
	}

	return EmptyCell
}
