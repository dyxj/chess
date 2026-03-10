package integration

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/dyxj/chess/pkg/engine"
	"github.com/dyxj/chess/pkg/game"
	"github.com/dyxj/chess/pkg/room"
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

	// After white connects, both players receive room ready and the initial RoundResult.
	w0, ok := <-wEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeRoomReady, w0.EventType)

	b0, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeRoomReady, b0.EventType)

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

	// After white connects, both players receive room ready and the initial RoundResult.
	w0, ok := <-wEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeRoomReady, w0.EventType)

	b0, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeRoomReady, b0.EventType)

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

	// After white connects, both players receive room ready and the initial RoundResult.
	w0, ok := <-wEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeRoomReady, w0.EventType)

	b0, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeRoomReady, b0.EventType)

	w1, ok := <-wEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeRoundResult, w1.EventType)

	b2, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeRoundResult, b2.EventType)

	// Game started
	moves := quickestCheckmate()
	mConn := wConn
	resultsW := make([]room.EventPartial, 0, len(moves))
	resultsB := make([]room.EventPartial, 0, len(moves))
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
		require.Equal(t, room.EventTypeRoundResult, wm.EventType, fmt.Sprintf("unexpected event type for move %d", i))
		resultsW = append(resultsW, wm)
		bm, okbm := <-bEventChan
		require.True(t, okbm, fmt.Sprintf("failed to receive event for move %d", i))
		require.Equal(t, room.EventTypeRoundResult, bm.EventType, fmt.Sprintf("unexpected event type for move %d", i))
		resultsB = append(resultsB, bm)
	}

	rrW, err := extractRoundResult(resultsW[len(moves)-1])
	require.NoError(t, err)
	assert.True(t, rrW.State.IsGameOver())
	assert.Equal(t, game.StateCheckmate, rrW.State)

	rrB, err := extractRoundResult(resultsB[len(moves)-1])
	require.NoError(t, err)
	assert.True(t, rrB.State.IsGameOver())
	assert.Equal(t, game.StateCheckmate, rrB.State)
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

	// After white connects, both players receive room ready and the initial RoundResult.
	w0, ok := <-wEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeRoomReady, w0.EventType)

	b0, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeRoomReady, b0.EventType)

	w1, ok := <-wEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeRoundResult, w1.EventType)

	b2, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeRoundResult, b2.EventType)

	// Game started
	moves := quickestStalemate()
	mConn := wConn
	resultsW := make([]room.EventPartial, 0, len(moves))
	resultsB := make([]room.EventPartial, 0, len(moves))
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
		require.Equal(t, room.EventTypeRoundResult, wm.EventType, fmt.Sprintf("unexpected event type for move %d", i))
		resultsW = append(resultsW, wm)
		bm, okbm := <-bEventChan
		require.True(t, okbm, fmt.Sprintf("failed to receive event for move %d", i))
		require.Equal(t, room.EventTypeRoundResult, bm.EventType, fmt.Sprintf("unexpected event type for move %d", i))
		resultsB = append(resultsB, bm)
	}

	rrW, err := extractRoundResult(resultsW[len(moves)-1])
	require.NoError(t, err)
	assert.True(t, rrW.State.IsGameOver())
	assert.Equal(t, game.StateStalemate, rrW.State)

	rrB, err := extractRoundResult(resultsB[len(moves)-1])
	require.NoError(t, err)
	assert.True(t, rrB.State.IsGameOver())
	assert.Equal(t, game.StateStalemate, rrB.State)
}

func TestRoomConnectHandler_TerminalError_InvalidToken(t *testing.T) {
	go testx.SetTimeout(t.Context(), 10*time.Second)

	logger := testx.GlobalEnv().Logger()
	testSvr := testx.GlobalEnv().HTTTPTestServer()

	eventChan, conn, err := websocketDialAndListen(
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), "invalid-token"),
		logger,
	)
	require.NoError(t, err)
	defer conn.Close()

	event, ok := <-eventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeError, event.EventType)

	var payload room.EventErrorPayload
	err = json.Unmarshal(event.Payload, &payload)
	require.NoError(t, err)
	require.True(t, payload.Terminal)
	require.Equal(t, room.ErrCodeInvalidToken, payload.Code)
	require.Equal(t, room.ErrInvalidToken.Error(), payload.Error)

	// Connection should close after terminal error
	_, ok = <-eventChan
	require.False(t, ok)
}

func TestRoomConnectHandler_TerminalError_ColorOccupied(t *testing.T) {
	go testx.SetTimeout(t.Context(), 10*time.Second)

	logger := testx.GlobalEnv().Logger()
	testSvr := testx.GlobalEnv().HTTTPTestServer()
	c := testx.GlobalEnv().RoomCoordinator()

	r, err := c.CreateRoom()
	require.NoError(t, err)

	// Issue two tokens for the same color while room is still waiting
	w1Token, err := c.IssueTicketToken(r.Code, "white player 1", engine.White)
	require.NoError(t, err)
	w2Token, err := c.IssueTicketToken(r.Code, "white player 2", engine.White)
	require.NoError(t, err)
	logger.Printf("code: %s, w1Token: %s, w2Token: %s\n", r.Code, w1Token, w2Token)

	// Connect w1: occupies the white color slot, room still waiting for black
	w1EventChan, w1Conn, err := websocketDialAndListen(
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), w1Token),
		logger,
	)
	require.NoError(t, err)
	defer w1Conn.Close()

	w1Msg, ok := <-w1EventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeMessage, w1Msg.EventType)

	// Connect w2: white slot already taken, room still StatusWaiting
	w2EventChan, w2Conn, err := websocketDialAndListen(
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), w2Token),
		logger,
	)
	require.NoError(t, err)
	defer w2Conn.Close()

	w2Event, ok := <-w2EventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeError, w2Event.EventType)

	var payload room.EventErrorPayload
	err = json.Unmarshal(w2Event.Payload, &payload)
	require.NoError(t, err)
	require.True(t, payload.Terminal)
	require.Equal(t, room.ErrCodeColorOccupied, payload.Code)
	require.Equal(t, room.ErrColorOccupied.Error(), payload.Error)

	// Connection should close after terminal error
	_, ok = <-w2EventChan
	require.False(t, ok)
}

func TestRoomConnectHandler_TerminalError_RoomFull(t *testing.T) {
	go testx.SetTimeout(t.Context(), 10*time.Second)

	logger := testx.GlobalEnv().Logger()
	testSvr := testx.GlobalEnv().HTTTPTestServer()
	c := testx.GlobalEnv().RoomCoordinator()

	// Issue all tokens before anyone connects (room still StatusWaiting)
	code, wToken, bToken, err := createRoomAndTokens(c)
	require.NoError(t, err)
	w2Token, err := c.IssueTicketToken(code, "white player 2", engine.White)
	require.NoError(t, err)
	logger.Printf("code: %s\n", code)

	// Connect black first: receives waiting message
	bEventChan, bConn, err := websocketDialAndListen(
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), bToken),
		logger,
	)
	require.NoError(t, err)
	defer bConn.Close()

	bWait, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeMessage, bWait.EventType)

	// Connect white: room transitions to StatusInProgress
	wEventChan, wConn, err := websocketDialAndListen(
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), wToken),
		logger,
	)
	require.NoError(t, err)
	defer wConn.Close()

	// Wait for both players to receive RoomReady (confirms room is StatusInProgress)
	wReady, ok := <-wEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeRoomReady, wReady.EventType)

	bReady, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeRoomReady, bReady.EventType)

	// Room is now StatusInProgress: w2 should get room_full terminal error
	w2EventChan, w2Conn, err := websocketDialAndListen(
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), w2Token),
		logger,
	)
	require.NoError(t, err)
	defer w2Conn.Close()

	w2Event, ok := <-w2EventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeError, w2Event.EventType)

	var payload room.EventErrorPayload
	err = json.Unmarshal(w2Event.Payload, &payload)
	require.NoError(t, err)
	require.True(t, payload.Terminal)
	require.Equal(t, room.ErrCodeRoomFull, payload.Code)
	require.Equal(t, room.ErrRoomFull.Error(), payload.Error)

	// Connection should close after terminal error
	_, ok = <-w2EventChan
	require.False(t, ok)
}

func TestRoomConnectHandler_GameError_WrongColor(t *testing.T) {
	go testx.SetTimeout(t.Context(), 10*time.Second)

	logger := testx.GlobalEnv().Logger()
	testSvr := testx.GlobalEnv().HTTTPTestServer()
	c := testx.GlobalEnv().RoomCoordinator()

	code, wToken, bToken, err := createRoomAndTokens(c)
	require.NoError(t, err)
	logger.Printf("code: %s\n", code)

	bEventChan, bConn, err := websocketDialAndListen(
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), bToken),
		logger,
	)
	require.NoError(t, err)
	defer bConn.Close()

	bWait, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeMessage, bWait.EventType)

	wEventChan, wConn, err := websocketDialAndListen(
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), wToken),
		logger,
	)
	require.NoError(t, err)
	defer wConn.Close()

	// Drain room ready and initial round result events
	_, ok = <-wEventChan
	require.True(t, ok)
	_, ok = <-bEventChan
	require.True(t, ok)
	_, ok = <-wEventChan
	require.True(t, ok)
	_, ok = <-bEventChan
	require.True(t, ok)

	// Black tries to move on white's turn
	err = writeActionMove(bConn, engine.Pawn, new(48), new(40))
	require.NoError(t, err)

	bErr, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeError, bErr.EventType)

	var errPayload room.EventErrorPayload
	err = json.Unmarshal(bErr.Payload, &errPayload)
	require.NoError(t, err)
	require.False(t, errPayload.Terminal)
	require.Equal(t, engine.ErrNotActiveColor.Error(), errPayload.Error)
}

func TestRoomConnectHandler_GameError_IllegalMove(t *testing.T) {
	go testx.SetTimeout(t.Context(), 10*time.Second)

	logger := testx.GlobalEnv().Logger()
	testSvr := testx.GlobalEnv().HTTTPTestServer()
	c := testx.GlobalEnv().RoomCoordinator()

	code, wToken, bToken, err := createRoomAndTokens(c)
	require.NoError(t, err)
	logger.Printf("code: %s\n", code)

	bEventChan, bConn, err := websocketDialAndListen(
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), bToken),
		logger,
	)
	require.NoError(t, err)
	defer bConn.Close()

	bWait, ok := <-bEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeMessage, bWait.EventType)

	wEventChan, wConn, err := websocketDialAndListen(
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), wToken),
		logger,
	)
	require.NoError(t, err)
	defer wConn.Close()

	// Drain room ready and initial round result events
	_, ok = <-wEventChan
	require.True(t, ok)
	_, ok = <-bEventChan
	require.True(t, ok)
	_, ok = <-wEventChan
	require.True(t, ok)
	_, ok = <-bEventChan
	require.True(t, ok)

	// White tries an illegal move (pawn backward from a2 to a1)
	err = writeActionMove(wConn, engine.Pawn, new(8), new(0))
	require.NoError(t, err)

	wErr, ok := <-wEventChan
	require.True(t, ok)
	require.Equal(t, room.EventTypeError, wErr.EventType)

	var errPayload room.EventErrorPayload
	err = json.Unmarshal(wErr.Payload, &errPayload)
	require.NoError(t, err)
	require.False(t, errPayload.Terminal)
	require.Equal(t, game.ErrIllegalMove.Error(), errPayload.Error)
}

func TestRoomConnectHandler_ActionPayload_ValidationErrors(t *testing.T) {
	//go testx.SetTimeout(t.Context(), 10*time.Second)
	tt := []struct {
		name    string
		moveMod func(r room.ActionMovePayload) room.ActionMovePayload
		errMsg  string
	}{
		{
			name: "from required",
			moveMod: func(r room.ActionMovePayload) room.ActionMovePayload {
				r.From = nil
				return r
			},
			errMsg: "from required",
		},
		{
			name: "to required",
			moveMod: func(r room.ActionMovePayload) room.ActionMovePayload {
				r.To = nil
				return r
			},
			errMsg: "to required",
		},
		{
			name: "from out of range -ve",
			moveMod: func(r room.ActionMovePayload) room.ActionMovePayload {
				r.From = new(-1)
				return r
			},
			errMsg: "from must be between 0 and 63",
		},
		{
			name: "from out of range >63",
			moveMod: func(r room.ActionMovePayload) room.ActionMovePayload {
				r.From = new(64)
				return r
			},
			errMsg: "from must be between 0 and 63",
		},
		{
			name: "to out of range -ve",
			moveMod: func(r room.ActionMovePayload) room.ActionMovePayload {
				r.To = new(-1)
				return r
			},
			errMsg: "to must be between 0 and 63",
		},
		{
			name: "to out of range >63",
			moveMod: func(r room.ActionMovePayload) room.ActionMovePayload {
				r.To = new(64)
				return r
			},
			errMsg: "to must be between 0 and 63",
		},
		{
			name: "invalid symbol",
			moveMod: func(r room.ActionMovePayload) room.ActionMovePayload {
				r.Symbol = 999
				return r
			},
			errMsg: "invalid symbol",
		},
		{
			name: "invalid symbol(no symbol)",
			moveMod: func(r room.ActionMovePayload) room.ActionMovePayload {
				r.Symbol = 0
				return r
			},
			errMsg: "invalid symbol",
		},
		{
			name: "invalid promotion symbol",
			moveMod: func(r room.ActionMovePayload) room.ActionMovePayload {
				r.Promotion = 999
				return r
			},
			errMsg: "invalid promotion symbol",
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

			// After white connects, both players receive room ready and the initial RoundResult.
			w0, ok := <-wEventChan
			require.True(t, ok)
			require.Equal(t, room.EventTypeRoomReady, w0.EventType)

			b0, ok := <-bEventChan
			require.True(t, ok)
			require.Equal(t, room.EventTypeRoomReady, b0.EventType)

			w1, ok := <-wEventChan
			require.True(t, ok)
			require.Equal(t, room.EventTypeRoundResult, w1.EventType)

			b2, ok := <-bEventChan
			require.True(t, ok)
			require.Equal(t, room.EventTypeRoundResult, b2.EventType)

			// Game started

			// Base Move
			move := room.ActionMovePayload{
				Symbol: engine.Pawn,
				From:   new(12),
				To:     new(20),
			}

			move = tc.moveMod(move)

			err = writeActionMove(wConn, move.Symbol, move.From, move.To, move.Promotion)
			require.NoError(t, err)
			wm, okwm := <-wEventChan
			require.True(t, okwm)
			require.Equal(t, room.EventTypeError, wm.EventType)

			var result room.EventErrorPayload
			err = json.Unmarshal(wm.Payload, &result)
			require.NoError(t, err)
			require.Equal(t, tc.errMsg, result.Error)
		})
	}
}
