package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

var dialer ws.Dialer

func websocketDialAndListen(
	url string,
	logger *log.Logger,
) (chan room.EventPartial, net.Conn, error) {

	conn, br, _, err := dialer.Dial(context.Background(), url)
	if err != nil {
		return nil, nil, err
	}
	eventChan := make(chan room.EventPartial, 10)

	// br is non-nil if the connection is buffered, so we need to use it for reading.
	// This could happen if the server sends some data immediately after the connection is established.
	var rw io.ReadWriter
	if br != nil {
		rw = struct {
			io.Reader
			io.Writer
		}{br, conn}
	} else {
		rw = conn
	}

	go func() {
		defer safe.Recover()()
		defer close(eventChan)
		if br != nil {
			defer func() {
				br.Reset(nil)
				ws.PutReader(br)
			}()
		}

		for {
			data, err := wsutil.ReadServerText(rw)
			if err != nil {
				logger.Printf("read error: %v", err)
				return
			}
			var event room.EventPartial
			if err = json.Unmarshal(data, &event); err != nil {
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

func quickestCheckmate() []ActionMove {
	return []ActionMove{
		NewActionMove(engine.Pawn, new(14), new(30)),  // 1. g4  (g2->g4)
		NewActionMove(engine.Pawn, new(52), new(44)),  // 1... e6 (e7->e6)
		NewActionMove(engine.Pawn, new(13), new(21)),  // 2. f3  (f2->f3)
		NewActionMove(engine.Queen, new(59), new(31)), // 2... Qh4# (d8->h4)
	}
}

func quickestStalemate() []ActionMove {
	return []ActionMove{
		NewActionMove(engine.Pawn, new(12), new(20)),  // 1. e3    (e2->e3)
		NewActionMove(engine.Pawn, new(48), new(32)),  // 1... a5  (a7->a5)
		NewActionMove(engine.Queen, new(3), new(39)),  // 2. Qh5   (d1->h5)
		NewActionMove(engine.Rook, new(56), new(40)),  // 2... Ra6 (a8->a6)
		NewActionMove(engine.Queen, new(39), new(32)), // 3. Qxa5  (h5->a5)
		NewActionMove(engine.Pawn, new(55), new(39)),  // 3... h5  (h7->h5)
		NewActionMove(engine.Queen, new(32), new(50)), // 4. Qxc7  (a5->c7)
		NewActionMove(engine.Rook, new(40), new(47)),  // 4... Rah6(a6->h6)
		NewActionMove(engine.Pawn, new(15), new(31)),  // 5. h4    (h2->h4)
		NewActionMove(engine.Pawn, new(53), new(45)),  // 5... f6  (f7->f6)
		NewActionMove(engine.Queen, new(50), new(51)), // 6. Qxd7+ (c7->d7)
		NewActionMove(engine.King, new(60), new(53)),  // 6... Kf7 (e8->f7)
		NewActionMove(engine.Queen, new(51), new(49)), // 7. Qxb7  (d7->b7)
		NewActionMove(engine.Queen, new(59), new(19)), // 7... Qd3 (d8->d3)
		NewActionMove(engine.Queen, new(49), new(57)), // 8. Qxb8  (b7->b8)
		NewActionMove(engine.Queen, new(19), new(55)), // 8... Qh7 (d3->h7)
		NewActionMove(engine.Queen, new(57), new(58)), // 9. Qxc8  (b8->c8)
		NewActionMove(engine.King, new(53), new(46)),  // 9... Kg6 (f7->g6)
		NewActionMove(engine.Queen, new(58), new(44)), // 10. Qe6  (c8->e6) Stalemate
	}
}
