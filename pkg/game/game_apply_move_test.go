package game

import (
	"math/rand/v2"
	"testing"

	"github.com/dyxj/chess/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestApplyMoveWithMock(t *testing.T) {
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
		b.EXPECT().ActiveColor().Return(engine.White)
		b.EXPECT().Piece(m.Color, m.Symbol, m.mbFrom()).Return(piece, true)
		b.EXPECT().GeneratePieceLegalMoves(piece).Return(legalMoves, nil)

		// Setup mock expectations for ApplyMove
		b.EXPECT().ApplyMove(engineMove).Return(nil)

		// Setup mock expectations for calculateGameState
		b.EXPECT().ActiveColor().Return(engine.Black) // After white's move
		b.EXPECT().HasLegalMoves(engine.Black).Return(true)
		b.EXPECT().MoveCount().Return(rand.IntN(10))
		b.EXPECT().GridRaw().Return([64]int{})
		b.EXPECT().ActiveColor().Return(engine.Black)

		_, err := g.ApplyMove(m)
		assert.NoError(t, err)

		assert.Equal(t, g.state, StateInProgress)
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
		b.EXPECT().ActiveColor().Return(engine.White)
		b.EXPECT().Piece(m.Color, m.Symbol, m.mbFrom()).Return(engine.Piece{}, false)

		_, err := g.ApplyMove(m)
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

		b.EXPECT().ActiveColor().Return(engine.White)
		b.EXPECT().Piece(m.Color, m.Symbol, m.mbFrom()).Return(piece, true)
		b.EXPECT().GeneratePieceLegalMoves(piece).Return(legalMoves, nil)

		_, err := g.ApplyMove(m)
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
		b.EXPECT().ActiveColor().Return(engine.White)
		b.EXPECT().Piece(m.Color, m.Symbol, m.mbFrom()).Return(piece, true)
		b.EXPECT().GeneratePieceLegalMoves(piece).Return(legalMoves, nil)

		// Setup mock expectations for ApplyMove
		b.EXPECT().ApplyMove(engineMove).Return(engine.ErrNotActiveColor)

		_, err := g.ApplyMove(m)
		assert.ErrorIs(t, err, engine.ErrNotActiveColor)

		assert.Equal(t, g.state, StateInProgress)
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
		b.EXPECT().ActiveColor().Return(engine.White)
		b.EXPECT().Piece(m.Color, m.Symbol, m.mbFrom()).Return(piece, true)
		b.EXPECT().GeneratePieceLegalMoves(piece).Return(legalMoves, nil)

		// Setup mock expectations for ApplyMove
		b.EXPECT().ApplyMove(engineMove).Return(nil)

		// Setup mock expectations for calculateGameState
		b.EXPECT().ActiveColor().Return(engine.Black) // After white's move

		// no moves and checked = checkmate
		b.EXPECT().HasLegalMoves(engine.Black).Return(false)
		b.EXPECT().IsCheck(engine.Black).Return(true)
		b.EXPECT().MoveCount().Return(rand.IntN(10))
		b.EXPECT().GridRaw().Return([64]int{})
		b.EXPECT().ActiveColor().Return(engine.Black)

		_, err := g.ApplyMove(m)
		assert.NoError(t, err)
		assert.Equal(t, g.state, StateCheckmate)
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
		b.EXPECT().ActiveColor().Return(engine.White)
		b.EXPECT().Piece(m.Color, m.Symbol, m.mbFrom()).Return(piece, true)
		b.EXPECT().GeneratePieceLegalMoves(piece).Return(legalMoves, nil)

		// Setup mock expectations for ApplyMove
		b.EXPECT().ApplyMove(engineMove).Return(nil)

		// Setup mock expectations for calculateGameState
		b.EXPECT().ActiveColor().Return(engine.Black) // After white's move

		// no moves and not checked = stalemate
		b.EXPECT().HasLegalMoves(engine.Black).Return(false)
		b.EXPECT().IsCheck(engine.Black).Return(false)
		b.EXPECT().MoveCount().Return(rand.IntN(10))
		b.EXPECT().GridRaw().Return([64]int{})
		b.EXPECT().ActiveColor().Return(engine.Black)

		_, err := g.ApplyMove(m)
		assert.NoError(t, err)
		assert.Equal(t, g.state, StateStalemate)
	})
}

func TestApplyMoveWithFileRank(t *testing.T) {
	t.Run("invalid input", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		b := NewMockBoard(ctrl)
		g := NewGame(b)

		tt := []struct {
			name  string
			input string
		}{
			{
				"length < 4",
				"a2b",
			},
			{
				"length > 4",
				"a2a1a3",
			},
			{
				"from invalid file",
				"i1a2",
			},
			{
				"from invalid rank",
				"a9a2",
			},
			{
				"from invalid file",
				"a1i2",
			},
			{
				"from invalid rank",
				"a1a9",
			},
		}

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				_, err := g.ApplyMoveWithFileRank(tc.input)
				assert.ErrorIs(t, err, ErrInvalidMove)
			})
		}
	})

	t.Run("successful conversion", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		b := NewMockBoard(ctrl)
		g := NewGame(b)

		b.EXPECT().ActiveColor().Return(engine.White).Times(1)
		b.EXPECT().Symbol(31).Return(engine.Pawn).Times(1)

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
		b.EXPECT().ActiveColor().Return(engine.White)
		b.EXPECT().Piece(m.Color, m.Symbol, m.mbFrom()).Return(piece, true)
		b.EXPECT().GeneratePieceLegalMoves(piece).Return(legalMoves, nil)

		// Setup mock expectations for ApplyMove
		b.EXPECT().ApplyMove(engineMove).Return(nil)

		// Setup mock expectations for calculateGameState
		b.EXPECT().ActiveColor().Return(engine.Black) // After white's move
		b.EXPECT().HasLegalMoves(engine.Black).Return(true)
		b.EXPECT().MoveCount().Return(rand.IntN(10))
		b.EXPECT().GridRaw().Return([64]int{})
		b.EXPECT().ActiveColor().Return(engine.Black)

		_, err := g.ApplyMoveWithFileRank("a2a3")
		assert.NoError(t, err)

		assert.Equal(t, g.state, StateInProgress)
	})
}

func TestApplyMove(t *testing.T) {
	t.Run("regular move", func(t *testing.T) {
		g := NewGame(engine.NewBoard())

		m := Move{
			Color:  engine.White,
			Symbol: engine.Knight,
			From:   1,
			To:     16,
		}

		_, err := g.ApplyMove(m)
		assert.NoError(t, err)

		assert.Equal(t, g.state, StateInProgress)
	})

	t.Run("promotion error(white), promotion not defined", func(t *testing.T) {
		board := engine.NewEmptyBoard(engine.White)
		// 84 uses mailbox representation
		piece := engine.NewPiece(engine.Pawn, engine.White, 84, true)
		err := board.LoadPieces([]engine.Piece{piece})
		require.NoError(t, err)

		g := NewGame(board)

		m := Move{
			Color:     engine.White,
			Symbol:    engine.Pawn,
			From:      51,
			To:        59,
			Promotion: 0,
		}

		_, err = g.ApplyMove(m)
		require.ErrorIs(t, err, ErrIllegalMove)
	})

	t.Run("promotion error(black), promotion not defined white", func(t *testing.T) {
		board := engine.NewEmptyBoard(engine.Black)
		// 84 uses mailbox representation
		piece := engine.NewPiece(engine.Pawn, engine.Black, 34, true)
		err := board.LoadPieces([]engine.Piece{piece})
		require.NoError(t, err)

		g := NewGame(board)

		m := Move{
			Color:     engine.Black,
			Symbol:    engine.Pawn,
			From:      11,
			To:        3,
			Promotion: 0,
		}

		_, err = g.ApplyMove(m)
		require.ErrorIs(t, err, ErrIllegalMove)
	})

	t.Run("promotion success(white)", func(t *testing.T) {
		tt := []struct {
			symbol engine.Symbol
		}{
			{engine.Queen},
			{engine.Rook},
			{engine.Bishop},
			{engine.Knight},
		}

		for _, tc := range tt {
			board := engine.NewEmptyBoard(engine.White)
			// 84 uses mailbox representation
			piece := engine.NewPiece(engine.Pawn, engine.White, 84, true)
			err := board.LoadPieces([]engine.Piece{piece})
			require.NoError(t, err)

			g := NewGame(board)

			m := Move{
				Color:     engine.White,
				Symbol:    engine.Pawn,
				From:      51,
				To:        59,
				Promotion: tc.symbol,
			}

			_, err = g.ApplyMove(m)
			assert.NoError(t, err)
		}
	})

	t.Run("promotion success(black)", func(t *testing.T) {
		tt := []struct {
			symbol engine.Symbol
		}{
			{engine.Queen},
			{engine.Rook},
			{engine.Bishop},
			{engine.Knight},
		}

		for _, tc := range tt {
			board := engine.NewEmptyBoard(engine.Black)
			// 84 uses mailbox representation
			piece := engine.NewPiece(engine.Pawn, engine.Black, 34, true)
			err := board.LoadPieces([]engine.Piece{piece})
			require.NoError(t, err)

			g := NewGame(board)

			m := Move{
				Color:     engine.Black,
				Symbol:    engine.Pawn,
				From:      11,
				To:        3,
				Promotion: tc.symbol,
			}

			_, err = g.ApplyMove(m)
			assert.NoError(t, err)
		}
	})
}
