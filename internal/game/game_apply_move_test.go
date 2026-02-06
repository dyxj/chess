package game

import (
	"testing"

	"github.com/dyxj/chess/internal/engine"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestApplyMove(t *testing.T) {
	t.Run("successful move", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		b := NewMockBoard(ctrl)
		g := NewGame(b)

		m := Move{
			Color:  engine.White,
			Symbol: engine.Pawn,
			From:   8,  // a2
			To:     16, // a3
		}

		// Mock piece at from position
		piece := engine.NewPiece(m.Symbol, m.Color, m.mbFrom())
		engineMove := engine.Move{
			Symbol: m.Symbol, Color: m.Color, From: m.mbFrom(), To: m.mbTo(),
		}
		legalMoves := []engine.Move{engineMove}

		// Setup mock expectations for validateAndConvertMove
		b.EXPECT().Piece(m.Color, m.Symbol, m.mbFrom()).Return(piece, true)
		b.EXPECT().GeneratePieceLegalMoves(piece).Return(legalMoves, nil)

		// Setup mock expectations for ApplyMove
		b.EXPECT().ApplyMove(engineMove).Return(nil)

		// Setup mock expectations for calculateGameState
		b.EXPECT().ActiveColor().Return(engine.Black) // After white's move
		b.EXPECT().HasLegalMoves(engine.Black).Return(true)

		err := g.ApplyMove(m)
		assert.NoError(t, err)

		assert.Equal(t, g.state, InProgress)
	})

	t.Run("piece not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		b := NewMockBoard(ctrl)
		g := NewGame(b)

		m := Move{
			Color:  engine.White,
			Symbol: engine.Pawn,
			From:   8,  // a2
			To:     16, // a3
		}

		// Mock piece not found
		b.EXPECT().Piece(m.Color, m.Symbol, m.mbFrom()).Return(engine.Piece{}, false)

		err := g.ApplyMove(m)
		assert.ErrorIs(t, err, engine.ErrPieceNotFound)
	})

	t.Run("illegal move", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		b := NewMockBoard(ctrl)
		g := NewGame(b)

		m := Move{
			Color:  engine.White,
			Symbol: engine.Pawn,
			From:   8,  // a2
			To:     32, // a5
		}

		// Mock piece exists but move is not legal
		piece := engine.NewPiece(m.Symbol, m.Color, m.mbFrom())
		engineMove := engine.Move{
			Symbol: m.Symbol, Color: m.Color, From: m.mbFrom(),
			To: engine.IndexToMailbox(24),
		}
		legalMoves := []engine.Move{engineMove}

		b.EXPECT().Piece(m.Color, m.Symbol, m.mbFrom()).Return(piece, true)
		b.EXPECT().GeneratePieceLegalMoves(piece).Return(legalMoves, nil)

		err := g.ApplyMove(m)
		assert.ErrorIs(t, err, ErrIllegalMove)
	})

	t.Run("failed to apply move", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		b := NewMockBoard(ctrl)
		g := NewGame(b)

		m := Move{
			Color:  engine.White,
			Symbol: engine.Pawn,
			From:   8,  // a2
			To:     16, // a3
		}

		// Mock piece at from position
		piece := engine.NewPiece(m.Symbol, m.Color, m.mbFrom())
		engineMove := engine.Move{
			Symbol: m.Symbol, Color: m.Color, From: m.mbFrom(), To: m.mbTo(),
		}
		legalMoves := []engine.Move{engineMove}

		// Setup mock expectations for validateAndConvertMove
		b.EXPECT().Piece(m.Color, m.Symbol, m.mbFrom()).Return(piece, true)
		b.EXPECT().GeneratePieceLegalMoves(piece).Return(legalMoves, nil)

		// Setup mock expectations for ApplyMove
		b.EXPECT().ApplyMove(engineMove).Return(engine.ErrNotActiveColor)

		err := g.ApplyMove(m)
		assert.ErrorIs(t, err, engine.ErrNotActiveColor)

		assert.Equal(t, g.state, InProgress)
	})

	t.Run("checkmate", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		b := NewMockBoard(ctrl)
		g := NewGame(b)

		m := Move{
			Color:  engine.White,
			Symbol: engine.Queen,
			From:   52,
			To:     60,
		}

		// Mock piece at from position
		piece := engine.NewPiece(m.Symbol, m.Color, m.mbFrom())
		engineMove := engine.Move{
			Symbol: m.Symbol, Color: m.Color, From: m.mbFrom(), To: m.mbTo(),
		}
		legalMoves := []engine.Move{engineMove}

		// Setup mock expectations for validateAndConvertMove
		b.EXPECT().Piece(m.Color, m.Symbol, m.mbFrom()).Return(piece, true)
		b.EXPECT().GeneratePieceLegalMoves(piece).Return(legalMoves, nil)

		// Setup mock expectations for ApplyMove
		b.EXPECT().ApplyMove(engineMove).Return(nil)

		// Setup mock expectations for calculateGameState
		b.EXPECT().ActiveColor().Return(engine.Black) // After white's move

		// no moves and checked = checkmate
		b.EXPECT().HasLegalMoves(engine.Black).Return(false)
		b.EXPECT().IsCheck(engine.Black).Return(true)

		err := g.ApplyMove(m)
		assert.NoError(t, err)
		assert.Equal(t, g.state, Checkmate)
	})

	t.Run("stalemate", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		b := NewMockBoard(ctrl)
		g := NewGame(b)

		m := Move{
			Color:  engine.White,
			Symbol: engine.Queen,
			From:   52,
			To:     60,
		}

		// Mock piece at from position
		piece := engine.NewPiece(m.Symbol, m.Color, m.mbFrom())
		engineMove := engine.Move{
			Symbol: m.Symbol, Color: m.Color, From: m.mbFrom(), To: m.mbTo(),
		}
		legalMoves := []engine.Move{engineMove}

		// Setup mock expectations for validateAndConvertMove
		b.EXPECT().Piece(m.Color, m.Symbol, m.mbFrom()).Return(piece, true)
		b.EXPECT().GeneratePieceLegalMoves(piece).Return(legalMoves, nil)

		// Setup mock expectations for ApplyMove
		b.EXPECT().ApplyMove(engineMove).Return(nil)

		// Setup mock expectations for calculateGameState
		b.EXPECT().ActiveColor().Return(engine.Black) // After white's move

		// no moves and not checked = stalemate
		b.EXPECT().HasLegalMoves(engine.Black).Return(false)
		b.EXPECT().IsCheck(engine.Black).Return(false)

		err := g.ApplyMove(m)
		assert.NoError(t, err)
		assert.Equal(t, g.state, Stalemate)
	})
}
