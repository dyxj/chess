package integration

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/dyxj/chess/internal/engine"
	"github.com/dyxj/chess/internal/game"
	"github.com/dyxj/chess/internal/room"
	"github.com/dyxj/chess/test/testx"
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
	require.Equal(t, room.EventTypeMessage, b1.EventType)

	var b1p room.EventMessagePayload
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
	require.Equal(t, room.EventTypeRoundResult, w1.EventType)

	b2, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeRoundResult, b2.EventType)

	// disconnect to resign
	wConn.Close()

	// After white disconnects, its channel should close.
	_, ok = <-wEventChan
	require.False(t, ok)

	// Black should receive a resign event.
	b3, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeResign, b3.EventType)

	var b3p room.EventResignPayload
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
	require.Equal(t, room.EventTypeMessage, b1.EventType)

	var b1p room.EventMessagePayload
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
	require.Equal(t, room.EventTypeRoundResult, w1.EventType)

	b2, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeRoundResult, b2.EventType)

	err = writeActionMove(wConn, engine.Pawn, new(11), new(19))
	require.NoError(t, err)

	// Both players should receive the result of the first move.
	firstResult := game.RoundResult{
		Count: 1,
		MoveResult: &game.MoveResult{
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
		State:       game.StateInProgress,
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
	secondResult := game.RoundResult{
		Count: 2,
		MoveResult: &game.MoveResult{
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
		State:       game.StateInProgress,
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
	require.Equal(t, room.EventTypeMessage, b1.EventType)

	var b1p room.EventMessagePayload
	err = json.Unmarshal(b1.Payload, &b1p)
	require.NoError(t, err)
	require.Equal(t, "Waiting for white player", b1p.Message)

	err = writeActionMove(bConn, engine.Pawn, new(48), new(40))
	require.NoError(t, err)

	// Wait for black player to get "waiting" message
	b2, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeMessage, b2.EventType)

	var b2p room.EventMessagePayload
	err = json.Unmarshal(b2.Payload, &b2p)
	require.NoError(t, err)
	require.Equal(t, "Discarding input as room is not ready", b2p.Message)
}
