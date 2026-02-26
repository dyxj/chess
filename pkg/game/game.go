package game

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/dyxj/chess/pkg/engine"
)

type Game struct {
	mu          sync.Mutex
	b           Board
	state       State
	winner      engine.Color
	CreatedTime time.Time
}

func NewGame(
	b Board,
) *Game {
	return &Game{
		b:           b,
		state:       StateInProgress,
		CreatedTime: time.Now(),
	}
}

func (g *Game) ApplyMove(m Move) (RoundResult, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.applyMove(m)
}

func (g *Game) UndoLastMove() bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.b.UndoLastMove()
}

func (g *Game) ActiveColor() engine.Color {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.b.ActiveColor()
}

func (g *Game) State() State {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.state
}

func (g *Game) GridRaw() [64]int {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.b.GridRaw()
}

func (g *Game) Pieces(c engine.Color) []engine.Piece {
	g.mu.Lock()
	defer g.mu.Unlock()

	pieces := g.b.Pieces(c)
	pp := make([]engine.Piece, len(pieces))
	for i := 0; i < len(g.b.Pieces(c)); i++ {
		pp[i] = pieces[i].WithPosition(
			engine.MailboxToIndex(pieces[i].Position()),
		)
	}
	return pp
}

var promotionNotationSymbol = map[string]engine.Symbol{
	"Q": engine.Queen,
	"R": engine.Rook,
	"B": engine.Bishop,
	"N": engine.Knight,
}

// ApplyMoveWithFileRank : format a2a3=N
// removes all spaces and converts to Move
// then calls ApplyMove
func (g *Game) ApplyMoveWithFileRank(move string) (RoundResult, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	move = strings.ReplaceAll(move, " ", "")

	split := strings.Split(move, "=")
	fromTo := split[0]
	if len(fromTo) != 4 {
		return RoundResult{}, fmt.Errorf("%w: input length is not equal 4", ErrInvalidMove)
	}

	promotion := engine.Symbol(0)
	if len(split) > 1 {
		pNotation := split[1]
		s, ok := promotionNotationSymbol[pNotation]
		if !ok {
			return RoundResult{}, fmt.Errorf("%w: invalid promotion piece", ErrInvalidMove)
		}
		promotion = s
	}

	if !g.isValidFile(fromTo[0]) ||
		!g.isValidRank(fromTo[1]) ||
		!g.isValidFile(fromTo[2]) ||
		!g.isValidRank(fromTo[3]) {
		return RoundResult{}, fmt.Errorf("%w: file or rank is out of range", ErrInvalidMove)
	}

	fromIndex := g.fileRankToIndex(fromTo[0], fromTo[1])
	m := Move{
		Color:     g.b.ActiveColor(),
		Symbol:    g.b.Symbol(engine.IndexToMailbox(fromIndex)),
		From:      fromIndex,
		To:        g.fileRankToIndex(fromTo[2], fromTo[3]),
		Promotion: promotion,
	}

	return g.applyMove(m)
}

func (g *Game) ForceDraw() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.canForceDraw() {
		g.state = StateDraw
		return nil
	}
	return ErrNotEligibleToForceDraw
}

func (g *Game) Winner() engine.Color {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.winner
}

func (g *Game) Symbol(pos int) engine.Symbol {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.b.Symbol(engine.IndexToMailbox(pos))
}

func (g *Game) Round() RoundResult {
	g.mu.Lock()
	defer g.mu.Unlock()

	var mr *MoveResult
	move, ok := g.b.LastMove()
	if ok {
		mr = new(fromEngine(move))
	}

	return RoundResult{
		Count:       g.b.MoveCount(),
		MoveResult:  mr,
		State:       g.state,
		Grid:        g.b.GridRaw(),
		ActiveColor: g.b.ActiveColor(),
	}
}

func (g *Game) Resign(color engine.Color) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.state.IsGameOver() {
		return fmt.Errorf("%w: game is already over", ErrInvalidMove)
	}

	if color == engine.White {
		g.state = StateWhiteResign
		g.winner = engine.Black
		return nil
	}

	g.state = StateBlackResign
	g.winner = engine.Black

	return nil
}

// ----- unexported ------- //
// makes it easier to check concurrent access

func (g *Game) applyMove(m Move) (RoundResult, error) {
	engineMove, err := g.validateAndConvertMove(m)
	if err != nil {
		return RoundResult{}, err
	}

	err = g.b.ApplyMove(engineMove)
	if err != nil {
		return RoundResult{}, err
	}

	g.state = g.calculateGameState()

	return RoundResult{
		Count:       g.b.MoveCount(),
		MoveResult:  new(fromEngine(engineMove)),
		State:       g.state,
		Grid:        g.b.GridRaw(),
		ActiveColor: g.b.ActiveColor(),
	}, nil
}

func (g *Game) validateAndConvertMove(m Move) (engine.Move, error) {
	if m.Color != g.b.ActiveColor() {
		return engine.Move{}, engine.ErrNotActiveColor
	}

	piece, ok := g.b.Piece(m.Color, m.Symbol, m.mbFrom())
	if !ok {
		return engine.Move{}, engine.ErrPieceNotFound
	}

	moves, err := g.b.GeneratePieceLegalMoves(piece)
	if err != nil {
		// board and piece out of sync, should panic due to programmer error
		panic(err)
	}

	moveIndex := slices.IndexFunc(moves, func(move engine.Move) bool {
		if move.From == m.mbFrom() && move.To == m.mbTo() {
			if move.Promotion == 0 {
				return true
			}
			return move.Promotion == m.Promotion
		}
		return false
	})
	if moveIndex == -1 {
		return engine.Move{}, ErrIllegalMove
	}

	return moves[moveIndex], nil
}

func (g *Game) canForceDraw() bool {
	return g.b.Is100MoveDraw() || g.b.Is3FoldDraw()
}

func (g *Game) isValidFile(r byte) bool {
	return r >= 'a' && r <= 'h'
}

func (g *Game) isValidRank(f byte) bool {
	return f >= '1' && f <= '8'
}

func (g *Game) fileRankToIndex(file, rank byte) int {
	fileIndex := int(file - 'a') // 'a'-'h' -> 0-7
	rankIndex := int(rank - '1') // '1'-'8' -> 0-7
	return rankIndex*8 + fileIndex
}
