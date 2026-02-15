package room

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/dyxj/chess/internal/engine"
	"github.com/dyxj/chess/internal/game"
	"github.com/dyxj/chess/pkg/safe"
	"github.com/dyxj/chess/pkg/websocketx"
	"go.uber.org/zap"
)

const createRoomMaxRetries = 5

type Coordinator struct {
	logger        *zap.Logger
	cache         *MemCache
	wsm           *websocketx.Manager
	ticketCache   *TicketCache
	tokenDuration time.Duration

	// [room.code]
	muRoomPubs     sync.RWMutex
	roomPublishers map[string]map[engine.Color]websocketPublisher
}

func NewCoordinator(logger *zap.Logger, tokenDuration time.Duration) *Coordinator {
	return &Coordinator{
		logger:         logger,
		cache:          NewMemCache(),
		wsm:            websocketx.NewManager(logger),
		ticketCache:    NewTicketCache(),
		tokenDuration:  tokenDuration,
		roomPublishers: make(map[string]map[engine.Color]websocketPublisher),
	}
}

// CreateRoom creates a new room and adds it to the cache.
// It retries for "createRoomMaxRetries" if the generated code already exists in the cache.
//
// Will add game/room configuration as input next
func (c *Coordinator) CreateRoom() (*Room, error) {
	room := NewEmptyRoom()
	retry := 0
	for {
		err := c.cache.Add(room)
		if err == nil {
			return room, nil
		}

		if !errors.Is(err, ErrCodeAlreadyExists) {
			return nil, fmt.Errorf("failed to create room due to unexpected error: %w", err)
		}

		if retry >= createRoomMaxRetries {
			return nil, fmt.Errorf("failed to create room after %d attempts: %w", retry, err)
		}

		c.logger.Warn("failed to add room to repo, retrying", zap.Error(err), zap.Int("retry", retry))
		retry++
		room.Code = generateCode()
	}
}

func (c *Coordinator) IssueTicketToken(code string, name string, color engine.Color) (string, error) {
	room, exist := c.cache.Find(code)
	if !exist {
		return "", ErrRoomNotFound
	}

	if room.Status() != StatusWaiting {
		return "", ErrRoomFull
	}

	err := room.IncrementTicket()
	if err != nil {
		return "", err
	}

	token := c.ticketCache.GenerateTicket(code, name, color, c.tokenDuration)
	time.AfterFunc(c.tokenDuration, func() {
		room.DecrementTicket()
	})

	return token, nil
}

func (c *Coordinator) ConnectWithToken(
	token string,
	w http.ResponseWriter,
	r *http.Request,
) error {
	ticket, valid := c.ticketCache.ConsumeTicket(token)
	if !valid {
		return ErrInvalidToken
	}

	room, exist := c.cache.Find(ticket.RoomCode)
	if !exist {
		return ErrRoomNotFound
	}

	if room.Status() != StatusWaiting {
		return ErrRoomFull
	}

	p := NewPlayer(ticket.Name)
	err := room.SetPlayer(ticket.Color, p)
	if err != nil {
		return err
	}

	conn, err := c.wsm.Open(p.ID.String(), w, r)
	if err != nil {
		room.RemovePlayer(ticket.Color)
		return err
	}
	defer c.wsm.Delete(conn.Key())

	defer func() {
		c.logger.Debug("closing websocket connection", zap.String("player", p.ID.String()))
		conn.CancelExistingOperations()
		c.logger.Debug("closed websocket connection", zap.String("player", p.ID.String()))
	}()

	c.runLoop(room, ticket.Color, conn)

	return nil
}

func (c *Coordinator) runLoop(
	room *Room,
	color engine.Color,
	ws websocketPublisherConsumer,
) {
	c.registerPublisher(room.Code, color, ws)
	defer c.unregisterPublisher(room.Code, color)

	roundResultChan := c.goConsumeLoop(room, color, ws)

	if room.HasBothPlayers() {
		room.SetStatus(StatusInProgress)
		room.signalReady()
	} else {
		xColor := color.Opposite()
		c.publishEventMessage(ws, fmt.Sprintf("Waiting for %v player", xColor))
	}
	<-room.readyChan

	c.publishEventRound(ws, room.Game.Round())

	c.goConsumeRoundResult(room, roundResultChan)

	<-room.gameOverChan

	c.logger.Debug("room finished", zap.String("room", room.Code))
}

// TODO error handling
//
//	Which errors should affect the process?
//	should check websocket.CloseStatus(err)
//	should check context, publish json currently doesn't provide context (context.Canceled)
func (c *Coordinator) publishEventMessage(p websocketPublisher, msg string) {
	e := NewEventMessage(msg)
	err := p.PublishJson(e)
	if err != nil {
		c.logger.Error("failed to publish WaitingForPlayer", zap.Error(err))
	}
}

// TODO error handling
//
//	Which errors should affect the process?
//	should check websocket.CloseStatus(err)
//	should check context, publish json currently doesn't provide context (context.Canceled)
func (c *Coordinator) publishEventRound(p websocketPublisher, round game.RoundResult) {
	err := p.PublishJson(round)
	if err != nil {
		c.logger.Error("failed to publish RoundResult", zap.Error(err))
	}
}

// TODO error handling
func (c *Coordinator) publishEventError(p websocketPublisher, eErr error) {
	e := NewEventError(eErr)
	err := p.PublishJson(e)
	if err != nil {
		c.logger.Error("failed to publish error event", zap.Any("eErr", eErr), zap.Error(err))
	}
}

func (c *Coordinator) registerPublisher(roomCode string, color engine.Color, pub websocketPublisher) {
	c.muRoomPubs.Lock()
	defer c.muRoomPubs.Unlock()

	pubs, ok := c.roomPublishers[roomCode]
	if !ok {
		pubs = make(map[engine.Color]websocketPublisher)
		c.roomPublishers[roomCode] = pubs
	}
	// overwriting existing is expected in case of reconnection
	pubs[color] = pub
}

func (c *Coordinator) unregisterPublisher(roomCode string, color engine.Color) {
	c.muRoomPubs.Lock()
	defer c.muRoomPubs.Unlock()

	if _, ok := c.roomPublishers[roomCode]; !ok {
		return
	}
	delete(c.roomPublishers[roomCode], color)
	if len(c.roomPublishers[roomCode]) == 0 {
		delete(c.roomPublishers, roomCode)
	}
}

func (c *Coordinator) goConsumeRoundResult(room *Room, roundResultChan <-chan game.RoundResult) {
	pubs, exist := c.roomPublishers[room.Code]
	if !exist {
		return
	}
	totalPubs := len(pubs)

	safe.GoWithLog(
		func() {
			for roundResult := range roundResultChan {
				wg := &sync.WaitGroup{}
				wg.Add(totalPubs)
				for color, pub := range pubs {
					// use concurrent publish to provide fairness
					// though it doesn't matter as much for chess
					// ranging over map provides random order
					safe.GoWithLog(
						func() {
							defer wg.Done()
							// TODO why did it keep broadcasting same result on publish failure
							err := pub.PublishJson(roundResult)
							if err != nil {
								c.logger.Error("failed to broadcast",
									zap.String("room", room.Code),
									zap.String("player", pub.Key()),
									zap.String("color", color.String()),
									zap.Error(err))
							}
						},
						c.logger, "broadcast panic",
					)
				}
				wg.Wait()
				if roundResult.State.IsGameOver() {
					room.signalGameOver()
					return
				}
			}
		},
		c.logger,
		"consume round result panic",
	)

}

// TODO error here needs to be propagated. In some cases the socket closes, it is crucial to notify parent
func (c *Coordinator) goConsumeLoop(
	room *Room,
	color engine.Color,
	ws websocketPublisherConsumer,
) <-chan game.RoundResult {

	roundResultChan := make(chan game.RoundResult)

	safe.GoWithLog(
		func() {
			defer close(roundResultChan)
			defer c.logger.Debug(
				"consume loop exited",
				zap.String("room", room.Code),
				zap.String("player", ws.Key()))
			for {
				c.logger.Debug("waiting for next message",
					zap.String("room", room.Code),
					zap.String("player", ws.Key()))
				var partial ActionPartial
				err := ws.ConsumeJson(&partial)
				if err != nil {
					status := websocket.CloseStatus(err)
					if status > 0 {
						c.logger.Info("websocket closed", zap.Any("close status", status))
						return
					}
					if errors.Is(err, context.Canceled) {
						c.logger.Debug("consume context canceled")
						return
					}
					c.logger.Error("failed to consume", zap.Error(err))
					return
				}

				switch partial.Type {
				case ActionTypeMove:
					err = c.consumeAction(room, color, partial.Payload, roundResultChan, ws)
					if err != nil {
						return
					}
				}
			}
		},
		c.logger,
		"consume loop panic",
	)

	return roundResultChan
}

func (c *Coordinator) consumeAction(
	room *Room,
	color engine.Color,
	data json.RawMessage,
	roundResultChan chan<- game.RoundResult,
	pub websocketPublisher,
) error {
	var payload ActionMovePayload
	err := json.Unmarshal(data, &payload)
	if err != nil {
		c.wsm.Close(
			pub.Key(),
			websocket.StatusInvalidFramePayloadData,
			"failed to unmarshal action payload",
		)
		return err
	}
	c.logger.Debug("payload", zap.Any("payload", payload))

	if room.Status() == StatusInProgress {
		c.logger.Debug("processing move action", zap.Any("payload", payload))
		return c.processMoveAction(room, color, payload, roundResultChan, pub)
	}

	if room.Status() == StatusWaiting {
		c.logger.Debug("discarding input due to waiting state")
		c.publishEventMessage(pub, fmt.Sprintf("Discarding input as room is not ready"))
		return nil
	}

	// Status Completed
	c.logger.Debug("discarding input due to completed state")
	c.publishEventMessage(pub, fmt.Sprintf("Discarding input as room is completed"))

	return nil
}

func (c *Coordinator) processMoveAction(
	room *Room,
	color engine.Color,
	payload ActionMovePayload,
	roundResultChan chan<- game.RoundResult,
	pub websocketPublisher,
) error {
	if err := payload.Validate(); err != nil {
		c.publishEventError(pub, err)
		return nil
	}

	result, err := room.Game.ApplyMove(
		payload.ToMove(color),
	)
	if err != nil {
		c.publishEventError(pub, err)
		return nil
	}

	roundResultChan <- result

	return nil
}

type websocketPublisher interface {
	Key() string
	PublishJson(v any) error
}

type websocketConsumer interface {
	Key() string
	ConsumeJson(v any) error
}

type websocketPublisherConsumer interface {
	websocketPublisher
	websocketConsumer
}
