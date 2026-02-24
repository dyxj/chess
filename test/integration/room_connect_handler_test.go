package integration

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/dyxj/chess/pkg/engine"
	game2 "github.com/dyxj/chess/pkg/game"
	room2 "github.com/dyxj/chess/pkg/room"
	"github.com/dyxj/chess/test/testx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const connectURLFormat = "ws://%s/room/connect?token=%s"

func TestRoomConnectHandler_Disconnect(t *testing.T) {
	go testx.SetTimeout(t.Context(), 10*time.Second)

	logger := testx.GlobalEnv().Logger()

	testSvr := testx.GlobalEnv().HTTTPTestServer()

	c := testx.GlobalEnv().RoomCoordinator()

	code, wToken, bToken, err := createRoomAndTokens(c)
	require.NoError(t, err)
	logger.Printf("code: %s, wToken: %s, bToken: %s\n", code, wToken, bToken)

	bEventChan, bConn, err := websocketDialAndListen(
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), bToken),
		logger,
	)
	require.NoError(t, err)
	defer bConn.Close()

	// Wait for black player to get "waiting" message
	b1, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room2.EventTypeMessage, b1.EventType)

	var b1p room2.EventMessagePayload
	err = json.Unmarshal(b1.Payload, &b1p)
	require.NoError(t, err)
	require.Equal(t, "Waiting for white player", b1p.Message)

	wEventChan, wConn, err := websocketDialAndListen(
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), wToken),
		logger,
	)
	require.NoError(t, err)
	defer wConn.Close()

	// After white connects, both players receive the initial RoundResult.
	w1, ok := <-wEventChan
	require.True(t, ok)
	require.Equal(t, room2.EventTypeRoundResult, w1.EventType)

	b2, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room2.EventTypeRoundResult, b2.EventType)

	// disconnect to resign
	wConn.Close()

	// After white disconnects, its channel should close.
	_, ok = <-wEventChan
	require.False(t, ok)

	// Black should receive a resign event.
	b3, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room2.EventTypeResign, b3.EventType)

	var b3p room2.EventResignPayload
	err = json.Unmarshal(b3.Payload, &b3p)
	require.NoError(t, err)
	require.Equal(t, engine.Black, b3p.Winner)
	require.Equal(t, engine.White, b3p.Resigner)
}

func TestRoomConnectHandler_SendActionMove(t *testing.T) {
	go testx.SetTimeout(t.Context(), 10*time.Second)

	logger := testx.GlobalEnv().Logger()

	testSvr := testx.GlobalEnv().HTTTPTestServer()

	c := testx.GlobalEnv().RoomCoordinator()

	code, wToken, bToken, err := createRoomAndTokens(c)
	require.NoError(t, err)
	logger.Printf("code: %s, wToken: %s, bToken: %s\n", code, wToken, bToken)

	bEventChan, bConn, err := websocketDialAndListen(
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), bToken),
		logger,
	)
	require.NoError(t, err)
	defer bConn.Close()

	// Wait for black player to get "waiting" message
	b1, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room2.EventTypeMessage, b1.EventType)

	var b1p room2.EventMessagePayload
	err = json.Unmarshal(b1.Payload, &b1p)
	require.NoError(t, err)
	require.Equal(t, "Waiting for white player", b1p.Message)

	wEventChan, wConn, err := websocketDialAndListen(
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), wToken),
		logger,
	)
	require.NoError(t, err)
	defer wConn.Close()

	// After white connects, both players receive the initial RoundResult.
	w1, ok := <-wEventChan
	require.True(t, ok)
	require.Equal(t, room2.EventTypeRoundResult, w1.EventType)

	b2, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room2.EventTypeRoundResult, b2.EventType)

	err = writeActionMove(wConn, engine.Pawn, new(11), new(19))
	require.NoError(t, err)

	// Both players should receive the result of the first move.
	firstResult := game2.RoundResult{
		Count: 1,
		MoveResult: &game2.MoveResult{
			Color:       engine.White,
			Symbol:      engine.Pawn,
			From:        11,
			To:          19,
			IsCastling:  false,
			RookFrom:    0,
			RookTo:      0,
			Captured:    0,
			Promotion:   0,
			IsEnPassant: false,
		},
		State:       game2.StateInProgress,
		Grid:        [64]int{4, 2, 3, 5, 6, 3, 2, 4, 1, 1, 1, 0, 1, 1, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, -1, -1, -1, -1, -1, -1, -1, -1, -4, -2, -3, -5, -6, -3, -2, -4},
		ActiveColor: engine.Black,
	}

	w2, ok := <-wEventChan
	require.True(t, ok)
	err = validateRoundResult(w2, firstResult)
	require.NoError(t, err)

	b3, ok := <-bEventChan
	require.True(t, ok)
	err = validateRoundResult(b3, firstResult)
	require.NoError(t, err)

	err = writeActionMove(bConn, engine.Pawn, new(48), new(40))
	require.NoError(t, err)

	// Both players should receive the result of the second move.
	secondResult := game2.RoundResult{
		Count: 2,
		MoveResult: &game2.MoveResult{
			Color:       engine.Black,
			Symbol:      engine.Pawn,
			From:        48,
			To:          40,
			IsCastling:  false,
			RookFrom:    0,
			RookTo:      0,
			Captured:    0,
			Promotion:   0,
			IsEnPassant: false,
		},
		State:       game2.StateInProgress,
		Grid:        [64]int{4, 2, 3, 5, 6, 3, 2, 4, 1, 1, 1, 0, 1, 1, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, -1, 0, 0, 0, 0, 0, 0, 0, 0, -1, -1, -1, -1, -1, -1, -1, -4, -2, -3, -5, -6, -3, -2, -4},
		ActiveColor: engine.White,
	}

	w3, ok := <-wEventChan
	require.True(t, ok)
	err = validateRoundResult(w3, secondResult)
	require.NoError(t, err)

	b4, ok := <-bEventChan
	require.True(t, ok)
	err = validateRoundResult(b4, secondResult)
	require.NoError(t, err)
}

func TestRoomConnectHandler_SendActionMove_Discard(t *testing.T) {
	go testx.SetTimeout(t.Context(), 10*time.Second)

	logger := testx.GlobalEnv().Logger()

	testSvr := testx.GlobalEnv().HTTTPTestServer()

	c := testx.GlobalEnv().RoomCoordinator()

	code, wToken, bToken, err := createRoomAndTokens(c)
	require.NoError(t, err)
	logger.Printf("code: %s, wToken: %s, bToken: %s\n", code, wToken, bToken)

	bEventChan, bConn, err := websocketDialAndListen(
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), bToken),
		logger,
	)
	require.NoError(t, err)
	defer bConn.Close()

	// Wait for black player to get "waiting" message
	b1, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room2.EventTypeMessage, b1.EventType)

	var b1p room2.EventMessagePayload
	err = json.Unmarshal(b1.Payload, &b1p)
	require.NoError(t, err)
	require.Equal(t, "Waiting for white player", b1p.Message)

	err = writeActionMove(bConn, engine.Pawn, new(48), new(40))
	require.NoError(t, err)

	// Wait for black player to get "waiting" message
	b2, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room2.EventTypeMessage, b2.EventType)

	var b2p room2.EventMessagePayload
	err = json.Unmarshal(b2.Payload, &b2p)
	require.NoError(t, err)
	require.Equal(t, "Discarding input as room is not ready", b2p.Message)
}

func TestRoomConnectHandler_Checkmate(t *testing.T) {
	go testx.SetTimeout(t.Context(), 10*time.Second)

	logger := testx.GlobalEnv().Logger()

	testSvr := testx.GlobalEnv().HTTTPTestServer()

	c := testx.GlobalEnv().RoomCoordinator()

	code, wToken, bToken, err := createRoomAndTokens(c)
	require.NoError(t, err)
	logger.Printf("code: %s, wToken: %s, bToken: %s\n", code, wToken, bToken)

	bEventChan, bConn, err := websocketDialAndListen(
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), bToken),
		logger,
	)
	require.NoError(t, err)
	defer bConn.Close()

	// Wait for black player to get "waiting" message
	b1, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room2.EventTypeMessage, b1.EventType)

	var b1p room2.EventMessagePayload
	err = json.Unmarshal(b1.Payload, &b1p)
	require.NoError(t, err)
	require.Equal(t, "Waiting for white player", b1p.Message)

	wEventChan, wConn, err := websocketDialAndListen(
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), wToken),
		logger,
	)
	require.NoError(t, err)
	defer wConn.Close()

	// After white connects, both players receive the initial RoundResult.
	w1, ok := <-wEventChan
	require.True(t, ok)
	require.Equal(t, room2.EventTypeRoundResult, w1.EventType)

	b2, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room2.EventTypeRoundResult, b2.EventType)

	// Game started
	moves := quickestCheckmate()
	mConn := wConn
	resultsW := make([]room2.EventPartial, 0, len(moves))
	resultsB := make([]room2.EventPartial, 0, len(moves))
	for i, move := range moves {
		err := writeActionMove(mConn, move.Payload.Symbol, move.Payload.From, move.Payload.To)
		require.NoError(t, err, fmt.Sprintf("move %d failed", i))
		if mConn == wConn {
			mConn = bConn
		} else {
			mConn = wConn
		}
		wm, okwm := <-wEventChan
		require.True(t, okwm, fmt.Sprintf("failed to receive event for move %d", i))
		require.Equal(t, room2.EventTypeRoundResult, wm.EventType, fmt.Sprintf("unexpected event type for move %d", i))
		resultsW = append(resultsW, wm)
		bm, okbm := <-bEventChan
		require.True(t, okbm, fmt.Sprintf("failed to receive event for move %d", i))
		require.Equal(t, room2.EventTypeRoundResult, bm.EventType, fmt.Sprintf("unexpected event type for move %d", i))
		resultsB = append(resultsB, bm)
	}

	rrW, err := extractRoundResult(resultsW[len(moves)-1])
	require.NoError(t, err)
	assert.True(t, rrW.State.IsGameOver())
	assert.Equal(t, game2.StateCheckmate, rrW.State)

	rrB, err := extractRoundResult(resultsB[len(moves)-1])
	require.NoError(t, err)
	assert.True(t, rrB.State.IsGameOver())
	assert.Equal(t, game2.StateCheckmate, rrB.State)
}

func TestRoomConnectHandler_Stalemate(t *testing.T) {
	go testx.SetTimeout(t.Context(), 10*time.Second)

	logger := testx.GlobalEnv().Logger()

	testSvr := testx.GlobalEnv().HTTTPTestServer()

	c := testx.GlobalEnv().RoomCoordinator()

	code, wToken, bToken, err := createRoomAndTokens(c)
	require.NoError(t, err)
	logger.Printf("code: %s, wToken: %s, bToken: %s\n", code, wToken, bToken)

	bEventChan, bConn, err := websocketDialAndListen(
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), bToken),
		logger,
	)
	require.NoError(t, err)
	defer bConn.Close()

	// Wait for black player to get "waiting" message
	b1, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room2.EventTypeMessage, b1.EventType)

	var b1p room2.EventMessagePayload
	err = json.Unmarshal(b1.Payload, &b1p)
	require.NoError(t, err)
	require.Equal(t, "Waiting for white player", b1p.Message)

	wEventChan, wConn, err := websocketDialAndListen(
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), wToken),
		logger,
	)
	require.NoError(t, err)
	defer wConn.Close()

	// After white connects, both players receive the initial RoundResult.
	w1, ok := <-wEventChan
	require.True(t, ok)
	require.Equal(t, room2.EventTypeRoundResult, w1.EventType)

	b2, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room2.EventTypeRoundResult, b2.EventType)

	// Game started
	moves := quickestStalemate()
	mConn := wConn
	resultsW := make([]room2.EventPartial, 0, len(moves))
	resultsB := make([]room2.EventPartial, 0, len(moves))
	for i, move := range moves {
		logger.Printf("move: %d", i)
		err := writeActionMove(mConn, move.Payload.Symbol, move.Payload.From, move.Payload.To)
		require.NoError(t, err, fmt.Sprintf("move %d failed", i))
		if mConn == wConn {
			mConn = bConn
		} else {
			mConn = wConn
		}
		wm, okwm := <-wEventChan
		require.True(t, okwm, fmt.Sprintf("failed to receive event for move %d", i))
		require.Equal(t, room2.EventTypeRoundResult, wm.EventType, fmt.Sprintf("unexpected event type for move %d", i))
		resultsW = append(resultsW, wm)
		bm, okbm := <-bEventChan
		require.True(t, okbm, fmt.Sprintf("failed to receive event for move %d", i))
		require.Equal(t, room2.EventTypeRoundResult, bm.EventType, fmt.Sprintf("unexpected event type for move %d", i))
		resultsB = append(resultsB, bm)
	}

	rrW, err := extractRoundResult(resultsW[len(moves)-1])
	require.NoError(t, err)
	assert.True(t, rrW.State.IsGameOver())
	assert.Equal(t, game2.StateStalemate, rrW.State)

	rrB, err := extractRoundResult(resultsB[len(moves)-1])
	require.NoError(t, err)
	assert.True(t, rrB.State.IsGameOver())
	assert.Equal(t, game2.StateStalemate, rrB.State)
}

func TestRoomConnectHandler_ActionPayload_ValidationErrors(t *testing.T) {
	go testx.SetTimeout(t.Context(), 10*time.Second)
	tt := []struct {
		name    string
		moveMod func(r room2.ActionMovePayload) room2.ActionMovePayload
		errMsg  string
	}{
		{
			name: "from required",
			moveMod: func(r room2.ActionMovePayload) room2.ActionMovePayload {
				r.From = nil
				return r
			},
			errMsg: "from required",
		},
		{
			name: "to required",
			moveMod: func(r room2.ActionMovePayload) room2.ActionMovePayload {
				r.To = nil
				return r
			},
			errMsg: "to required",
		},
		{
			name: "from out of range -ve",
			moveMod: func(r room2.ActionMovePayload) room2.ActionMovePayload {
				r.From = new(-1)
				return r
			},
			errMsg: "from must be between 0 and 63",
		},
		{
			name: "from out of range >63",
			moveMod: func(r room2.ActionMovePayload) room2.ActionMovePayload {
				r.From = new(64)
				return r
			},
			errMsg: "from must be between 0 and 63",
		},
		{
			name: "to out of range -ve",
			moveMod: func(r room2.ActionMovePayload) room2.ActionMovePayload {
				r.To = new(-1)
				return r
			},
			errMsg: "to must be between 0 and 63",
		},
		{
			name: "to out of range >63",
			moveMod: func(r room2.ActionMovePayload) room2.ActionMovePayload {
				r.To = new(64)
				return r
			},
			errMsg: "to must be between 0 and 63",
		},
		{
			name: "invalid symbol",
			moveMod: func(r room2.ActionMovePayload) room2.ActionMovePayload {
				r.Symbol = 999
				return r
			},
			errMsg: "invalid symbol",
		},
		{
			name: "invalid symbol(no symbol)",
			moveMod: func(r room2.ActionMovePayload) room2.ActionMovePayload {
				r.Symbol = 0
				return r
			},
			errMsg: "invalid symbol",
		},
	}

	logger := testx.GlobalEnv().Logger()

	testSvr := testx.GlobalEnv().HTTTPTestServer()

	c := testx.GlobalEnv().RoomCoordinator()

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			code, wToken, bToken, err := createRoomAndTokens(c)
			require.NoError(t, err)
			logger.Printf("code: %s, wToken: %s, bToken: %s\n", code, wToken, bToken)

			bEventChan, bConn, err := websocketDialAndListen(
				fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), bToken),
				logger,
			)
			require.NoError(t, err)
			defer bConn.Close()

			// Wait for black player to get "waiting" message
			b1, ok := <-bEventChan
			require.True(t, ok)
			require.Equal(t, room2.EventTypeMessage, b1.EventType)

			var b1p room2.EventMessagePayload
			err = json.Unmarshal(b1.Payload, &b1p)
			require.NoError(t, err)
			require.Equal(t, "Waiting for white player", b1p.Message)

			wEventChan, wConn, err := websocketDialAndListen(
				fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), wToken),
				logger,
			)
			require.NoError(t, err)
			defer wConn.Close()

			// After white connects, both players receive the initial RoundResult.
			w1, ok := <-wEventChan
			require.True(t, ok)
			require.Equal(t, room2.EventTypeRoundResult, w1.EventType)

			b2, ok := <-bEventChan
			require.True(t, ok)
			require.Equal(t, room2.EventTypeRoundResult, b2.EventType)

			// Game started

			// Base Move
			move := room2.ActionMovePayload{
				Symbol: engine.Pawn,
				From:   new(12),
				To:     new(20),
			}

			move = tc.moveMod(move)

			err = writeActionMove(wConn, move.Symbol, move.From, move.To)
			require.NoError(t, err)
			wm, okwm := <-wEventChan
			require.True(t, okwm)
			require.Equal(t, room2.EventTypeError, wm.EventType)

			var result room2.EventErrorPayload
			err = json.Unmarshal(wm.Payload, &result)
			require.NoError(t, err)
			require.Equal(t, tc.errMsg, result.Error)
		})
	}
}
