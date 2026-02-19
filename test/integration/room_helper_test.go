package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/dyxj/chess/internal/engine"
	"github.com/dyxj/chess/internal/game"
	"github.com/dyxj/chess/internal/room"
	"github.com/dyxj/chess/pkg/safe"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/go-cmp/cmp"
)

func websocketDialAndListen(
	url string,
	logger *log.Logger,
) (chan room.EventPartial, net.Conn, error) {

	conn, _, _, err := ws.Dialer{}.Dial(context.Background(), url)
	if err != nil {
		return nil, nil, err
	}

	eventChan := make(chan room.EventPartial, 100)

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

	return eventChan, conn, nil
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

func writeActionMove(conn net.Conn, symbol engine.Symbol, from *int, to *int) error {
	action := NewActionMove(symbol, from, to)
	data, err := json.Marshal(action)
	if err != nil {
		return err
	}

	err = wsutil.WriteClientText(conn, data)
	if err != nil {
		return err
	}
	return nil
}

func extractRoundResult(event room.EventPartial) (game.RoundResult, error) {
	var result game.RoundResult
	err := json.Unmarshal(event.Payload, &result)
	if err != nil {
		return game.RoundResult{}, err
	}
	return result, nil
}

type ActionMove struct {
	Type    room.ActionType        `json:"type"`
	Payload room.ActionMovePayload `json:"payload"`
}

func NewActionMove(symbol engine.Symbol, from *int, to *int) ActionMove {
	return ActionMove{
		Type: room.ActionTypeMove,
		Payload: room.ActionMovePayload{
			Symbol: symbol,
			From:   from,
			To:     to,
		},
	}
}

func validateRoundResult(event room.EventPartial, exp game.RoundResult) error {
	if event.EventType != room.EventTypeRoundResult {
		return fmt.Errorf("unexpected event type: %s", event.EventType)
	}
	result, err := extractRoundResult(event)
	if err != nil {
		return fmt.Errorf("failed to extract round result: %w", err)
	}

	if !cmp.Equal(exp, result,
		cmp.Transformer("DerefMoveResult", func(m *game.MoveResult) game.MoveResult {
			if m == nil {
				return game.MoveResult{}
			}
			return *m
		}),
	) {

		resultMr := game.MoveResult{}
		if result.MoveResult != nil {
			resultMr = *result.MoveResult
		}
		expMr := game.MoveResult{}
		if exp.MoveResult != nil {
			expMr = *exp.MoveResult
		}
		return fmt.Errorf("unexpected round result: %+v\n%+v\nexp: %+v\n%+v",
			result, resultMr, exp, expMr)
	}

	return nil
}
