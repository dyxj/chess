package game

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/dyxj/chess/internal/engine"
)

type Game struct {
	mu     sync.Mutex
	b      Board
	state  State
	winner engine.Color
}

func NewGame(
	b Board,
) *Game {
	return &Game{
		b:     b,
		state: InProgress,
	}
}

func (g *Game) ApplyMove(m Move) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	engineMove, err := g.validateAndConvertMove(m)
	if err != nil {
		return err
	}

	err = g.b.ApplyMove(engineMove)
	if err != nil {
		return err
	}

	g.state = g.calculateGameState()

	return nil
}

func (g *Game) validateAndConvertMove(m Move) (engine.Move, error) {
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
			return true
		}
		return false
	})
	if moveIndex == -1 {
		return engine.Move{}, ErrIllegalMove
	}

	return moves[moveIndex], nil
}

func (g *Game) UndoLastMove() bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.b.UndoLastMove()
}

func (g *Game) ActiveColor() engine.Color {
	return g.b.ActiveColor()
}

func (g *Game) State() State {
	return g.state
}

func (g *Game) GridRaw() [64]int {
	return g.b.GridRaw()
}

// ApplyMoveWithFileRank : format a2a3
// removes all spaces and converts to Move
// then calls ApplyMove
func (g *Game) ApplyMoveWithFileRank(move string) error {
	move = strings.ReplaceAll(move, " ", "")
	if len(move) != 4 {
		return fmt.Errorf("%w: input length is not equal 4", ErrIllegalMove)
	}

	if !g.isValidFile(move[0]) ||
		!g.isValidRank(move[1]) ||
		!g.isValidFile(move[2]) ||
		!g.isValidRank(move[3]) {
		return fmt.Errorf("%w: file or rank is out of range", ErrIllegalMove)
	}

	fromIndex := g.fileRankToIndex(move[0], move[1])
	m := Move{
		Color:  g.b.ActiveColor(),
		Symbol: g.b.Symbol(engine.IndexToMailbox(fromIndex)),
		From:   fromIndex,
		To:     g.fileRankToIndex(move[2], move[3]),
	}

	return g.ApplyMove(m)
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

func (g *Game) ForceDraw() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.canForceDraw() {
		g.state = Draw
		return nil
	}
	return ErrNotEligibleToForceDraw
}

func (g *Game) canForceDraw() bool {
	return g.b.Is100MoveDraw() || g.b.Is3FoldDraw()
}

func (g *Game) Winner() engine.Color {
	return g.winner
}
