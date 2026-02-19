package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/dyxj/chess/internal/engine"
	"github.com/dyxj/chess/internal/room"
	"github.com/dyxj/chess/pkg/safe"
	"github.com/dyxj/chess/test/testx"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const connectURLFormat = "ws://%s/room/connect?token=%s"

func TestRoomConnectHandlerDisconnect(t *testing.T) {
	logger := testx.GlobalEnv().Logger()

	testSvr := testx.GlobalEnv().HTTTPTestServer()

	memCache := testx.GlobalEnv().MemCache()
	t.Cleanup(func() {
		memCache.Clear()
	})

	c := testx.GlobalEnv().RoomCoordinator()

	_, wToken, bToken, err := createRoomAndTokens(c)
	require.NoError(t, err)

	ctx, cancelFunc := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancelFunc()

	bEventChan, err := dialAndListen(ctx,
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), bToken),
		logger,
		1*time.Second,
	)
	require.NoError(t, err)

	wEventChan, err := dialAndListen(ctx,
		fmt.Sprintf(connectURLFormat, testSvr.Listener.Addr().String(), wToken),
		logger,
		500*time.Millisecond,
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
}

func dialAndListen(ctx context.Context,
	url string,
	logger *log.Logger,
	closeAfter time.Duration,
) (chan room.EventPartial, error) {

	conn, _, _, err := ws.Dial(ctx, url)
	if err != nil {
		return nil, err
	}

	eventChan := make(chan room.EventPartial, 100)
	go func() {
		defer safe.Recover()
		defer func() {
			err := conn.Close()
			if err != nil {
				logger.Println(err)
			}
		}()
		select {
		case <-ctx.Done():
			logger.Println("closing due to context done")
			return
		case <-time.After(closeAfter):
			logger.Println("closing due to timeout")
			return
		}
	}()

	go func() {
		defer safe.Recover()
		defer close(eventChan)

		for {
			data, err := wsutil.ReadServerText(conn)
			if err != nil {
				logger.Printf("read error: %v", err)
				return
			}
			var event room.EventPartial
			err = json.Unmarshal(data, &event)
			if err != nil {
				logger.Printf("unmarshal error: %v", err)
				return
			}
			eventChan <- event
		}
	}()

	return eventChan, nil
}

func createRoomAndTokens(
	c *room.Coordinator,
) (code string, wToken string, bToken string, err error) {

	r, err := c.CreateRoom()
	if err != nil {
		return "", "", "", err
	}

	wToken, err = c.IssueTicketToken(r.Code, "white player", engine.White)
	if err != nil {
		return "", "", "", err
	}
	bToken, err = c.IssueTicketToken(r.Code, "black player", engine.Black)
	if err != nil {
		return "", "", "", err
	}

	return r.Code, wToken, bToken, nil
}
