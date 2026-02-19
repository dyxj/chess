package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/dyxj/chess/internal/engine"
	"github.com/dyxj/chess/internal/game"
	"github.com/dyxj/chess/internal/room"
	"github.com/dyxj/chess/test/testx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const connectURLFormat = "ws://%s/room/connect?token=%s"

func TestRoomConnectHandler_Disconnect(t *testing.T) {
	logger := testx.GlobalEnv().Logger()

	testSvr := testx.GlobalEnv().HTTTPTestServer()

	memCache := testx.GlobalEnv().MemCache()
	t.Cleanup(func() {
		memCache.Clear()
	})

	c := testx.GlobalEnv().RoomCoordinator()

	_, wToken, bToken, err := createRoomAndTokens(c)
	require.NoError(t, err)

	bctx, bcancelFunc := context.WithTimeout(t.Context(), 5*time.Second)
	defer bcancelFunc()

	bEventChan, _, bCloseDone, err := websocketDialAndListen(bctx,
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), bToken),
		logger,
	)
	require.NoError(t, err)

	<-time.After(100 * time.Millisecond)

	wctx, wcancelFunc := context.WithTimeout(t.Context(), 4*time.Second)
	defer wcancelFunc()
	wEventChan, _, wCloseDone, err := websocketDialAndListen(wctx,
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), wToken),
		logger,
	)
	require.NoError(t, err)

	<-time.After(50 * time.Millisecond)

	b1, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeMessage, b1.EventType)

	var b1p room.EventMessagePayload
	err = json.Unmarshal(b1.Payload, &b1p)
	require.NoError(t, err)
	require.Equal(t, "Waiting for white player", b1p.Message)

	w1, ok := <-wEventChan
	require.True(t, ok)
	assert.Equal(t, room.EventTypeRoundResult, w1.EventType)

	b2, ok := <-bEventChan
	require.True(t, ok)
	assert.Equal(t, room.EventTypeRoundResult, b2.EventType)

	// disconnect
	wcancelFunc()

	_, ok = <-wEventChan
	require.False(t, ok)

	b3, ok := <-bEventChan
	require.True(t, ok)
	assert.Equal(t, room.EventTypeResign, b3.EventType)

	var b3p room.EventResignPayload
	err = json.Unmarshal(b3.Payload, &b3p)
	require.NoError(t, err)
	assert.Equal(t, engine.Black, b3p.Winner)
	assert.Equal(t, engine.White, b3p.Resigner)

	<-bCloseDone
	<-wCloseDone
}

func TestRoomConnectHandler_SendAction(t *testing.T) {
	logger := testx.GlobalEnv().Logger()

	testSvr := testx.GlobalEnv().HTTTPTestServer()

	memCache := testx.GlobalEnv().MemCache()
	t.Cleanup(func() {
		memCache.Clear()
	})

	c := testx.GlobalEnv().RoomCoordinator()

	_, wToken, bToken, err := createRoomAndTokens(c)
	require.NoError(t, err)

	bctx, bcancelFunc := context.WithTimeout(t.Context(), 5*time.Second)
	defer bcancelFunc()

	bEventChan, bConn, bCloseDone, err := websocketDialAndListen(bctx,
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), bToken),
		logger,
	)
	require.NoError(t, err)

	<-time.After(100 * time.Millisecond)

	wctx, wcancelFunc := context.WithTimeout(t.Context(), 4*time.Second)
	defer wcancelFunc()
	wEventChan, wConn, wCloseDone, err := websocketDialAndListen(wctx,
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), wToken),
		logger,
	)
	require.NoError(t, err)

	<-time.After(50 * time.Millisecond)

	b1, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeMessage, b1.EventType)

	var b1p room.EventMessagePayload
	err = json.Unmarshal(b1.Payload, &b1p)
	require.NoError(t, err)
	require.Equal(t, "Waiting for white player", b1p.Message)

	err = writeActionMove(wConn, engine.Pawn, new(11), new(19))
	require.NoError(t, err)

	w1, ok := <-wEventChan
	require.True(t, ok)
	assert.Equal(t, room.EventTypeRoundResult, w1.EventType)

	b2, ok := <-bEventChan
	require.True(t, ok)
	assert.Equal(t, room.EventTypeRoundResult, b2.EventType)

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

	err = writeActionMove(bConn, engine.Pawn, new(48), new(40))
	require.NoError(t, err)

	w2, ok := <-wEventChan
	require.True(t, ok)
	err = validateRoundResult(w2, firstResult)
	require.NoError(t, err)

	b3, ok := <-bEventChan
	require.True(t, ok)
	err = validateRoundResult(b3, firstResult)
	require.NoError(t, err)

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

	bcancelFunc()
	wcancelFunc()

	<-bCloseDone
	<-wCloseDone
}
