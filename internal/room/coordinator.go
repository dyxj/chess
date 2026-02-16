package room

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dyxj/chess/internal/engine"
	"github.com/dyxj/chess/internal/game"
	"github.com/dyxj/chess/pkg/safe"
	"github.com/dyxj/chess/pkg/websocketx"
	"github.com/gobwas/ws"
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

	logger := c.logger.With(
		zap.String("player", p.ID.String()),
		zap.String("room", room.Code),
		zap.String("color", ticket.Color.String()),
	)

	defer func() {
		logger.Debug("closing websocket connection")
		conn.Cancel()
		err := conn.WriteCloseStatusCode(ws.StatusNormalClosure, "game over")
		if err != nil {
			logger.Info("failed to write close status", zap.Error(err))
		}
		err = conn.Close()
		if err != nil {
			logger.Info("failed to close websocket connection", zap.Error(err))
		}
		logger.Debug("closed websocket connection")
	}()

	err = c.runLoop(room, ticket.Color, conn, logger)
	if err != nil {
		return c.handleRunLoopError(err, room, ticket.Color)
	}

	return nil
}

func (c *Coordinator) runLoop(
	room *Room,
	color engine.Color,
	ws websocketPublisherConsumer,
	logger *zap.Logger,
) error {
	c.registerPublisher(room.Code, color, ws)
	defer c.unregisterPublisher(room.Code, color)

	roundResultChan, consumeErrChan := c.goConsumeLoop(room, color, ws, logger)

	if room.HasBothPlayers() {
		room.SetStatus(StatusInProgress)
		room.signalReady()
	} else {
		xColor := color.Opposite()
		err := c.publishEventMessage(ws, fmt.Sprintf("Waiting for %v player", xColor))
		if err != nil {
			return err
		}
	}
	<-room.readyChan

	err := c.publishEventRound(ws, room.Game.Round())
	if err != nil {
		return err
	}

	err = c.goConsumeRoundResult(room, roundResultChan, logger)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ws.Context().Done():
			logger.Debug("exiting run loop, websocket context closed")
			return nil
		case consumeErr, ok := <-consumeErrChan:
			if !ok {
				consumeErrChan = nil
			}
			logger.Debug("exiting run loop, due to consume error")
			return consumeErr
		case <-room.gameOverChan:
			logger.Debug("exiting run loop, due to game over")
			return nil
		}
	}
}

func (c *Coordinator) publishEventMessage(p websocketPublisher, msg string) error {
	e := NewEventMessage(msg)
	err := p.PublishJson(e)
	if err != nil {
		return err
	}
	return nil
}

func (c *Coordinator) publishEventRound(p websocketPublisher, round game.RoundResult) error {
	err := p.PublishJson(round)
	if err != nil {
		return err
	}
	return nil
}

func (c *Coordinator) publishEventError(p websocketPublisher, eErr error) error {
	e := NewEventError(eErr)
	err := p.PublishJson(e)
	if err != nil {
		return err
	}
	return nil
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

func (c *Coordinator) goConsumeRoundResult(
	room *Room,
	roundResultChan <-chan game.RoundResult,
	logger *zap.Logger,
) error {
	pubs, exist := c.roomPublishers[room.Code]
	if !exist {
		return fmt.Errorf("no publishers found for room %s", room.Code)
	}
	totalPubs := len(pubs)

	// TODO here, if publishing fail here something should be returned, same as consume loop
	go func() {
		defer safe.RecoverWithLog(logger, "goConsumeRoundResult")

		for roundResult := range roundResultChan {
			wg := &sync.WaitGroup{}
			wg.Add(totalPubs)
			for color, pub := range pubs {
				// use concurrent publish to provide fairness
				// though it doesn't matter as much for chess
				// ranging over map provides random order
				go func() {
					lg := c.logger.With(
						zap.String("room", room.Code),
						zap.String("color", color.String()),
					)
					defer safe.RecoverWithLog(lg, "goConsumeRoundResult:broadcast")
					defer wg.Done()

					err := pub.PublishJson(NewEventRound(roundResult))
					if err != nil {
						lg.Error("failed to broadcast", zap.Error(err))
						return
					}
				}()
			}

			wg.Wait()

			if roundResult.State.IsGameOver() {
				room.signalGameOver()
				return
			}
		}
	}()

	return nil
}

func (c *Coordinator) goConsumeLoop(
	room *Room,
	color engine.Color,
	ws websocketPublisherConsumer,
	logger *zap.Logger,
) (<-chan game.RoundResult, <-chan error) {

	roundResultChan := make(chan game.RoundResult)
	errChan := make(chan error)
	go func() {
		defer safe.RecoverWithLog(logger, "goConsumeLoop")
		defer logger.Debug("consume loop exited")
		defer close(roundResultChan)
		defer close(errChan)

		for {
			select {
			case <-ws.Context().Done():
				logger.Info("websocket context closed")
				return

			default:
				logger.Debug("waiting for next message")

				var partial ActionPartial
				err := ws.ConsumeJson(&partial)
				if err != nil {
					if websocketx.IsNetworkClosedError(err) {
						logger.Info("network closed")
					} else if wsErr, isErr := websocketx.IsWebSocketClosedError(err); isErr {
						logger.Info("websocket closed by client", zap.Error(wsErr))
					} else {
						logger.Error("failed to consume", zap.Error(err))
					}

					errChan <- err
					return
				}

				switch partial.Type {
				case ActionTypeMove:
					err = c.processAction(room, color, partial.Payload, roundResultChan, ws, logger)
					if err != nil {
						logger.Error("failed to process action", zap.Error(err))
						errChan <- err
						return
					}
				}
			}
		}
	}()

	return roundResultChan, errChan
}

func (c *Coordinator) processAction(
	room *Room,
	color engine.Color,
	data json.RawMessage,
	roundResultChan chan<- game.RoundResult,
	pub websocketPublisher,
	logger *zap.Logger,
) error {
	var payload ActionMovePayload
	err := json.Unmarshal(data, &payload)
	if err != nil {
		return err
	}
	logger.Debug("process action", zap.Any("payload", payload))

	if room.Status() == StatusInProgress {
		logger.Debug("processing move action", zap.Any("payload", payload))
		return c.processMoveAction(room, color, payload, roundResultChan, pub)
	}

	if room.Status() == StatusWaiting {
		logger.Debug("discarding input due to waiting state")
		pErr := c.publishEventMessage(pub, fmt.Sprintf("Discarding input as room is not ready"))
		if pErr != nil {
			return pErr
		}
		return nil
	}

	// Status Completed, small likelihood
	logger.Debug("discarding input due to completed state")
	pErr := c.publishEventMessage(pub, fmt.Sprintf("Discarding input as room is completed"))
	if pErr != nil {
		return pErr
	}

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
		pErr := c.publishEventError(pub, err)
		if pErr != nil {
			return pErr
		}
		return nil
	}

	result, err := room.Game.ApplyMove(
		payload.ToMove(color),
	)
	if err != nil {
		pErr := c.publishEventError(pub, err)
		if pErr != nil {
			return pErr
		}
		return nil
	}

	roundResultChan <- result

	return nil
}

func (c *Coordinator) handleRunLoopError(err error, room *Room, color engine.Color) error {
	if websocketx.IsNetworkClosedError(err) {
		return err
	}

	if wsErr, isErr := websocketx.IsWebSocketClosedError(err); isErr {
		c.resignAndNotifyOpponent(room, color)
		if wsErr.Code == ws.StatusNormalClosure {
			return nil
		}
		return err
	}

	return err
}

func (c *Coordinator) resignAndNotifyOpponent(room *Room, color engine.Color) {
	if room.Game.State().IsGameOver() {
		return
	}

	err := room.Game.Resign(color)
	if err != nil {
		c.logger.Error("failed to resign game after websocket closed",
			zap.String("room", room.Code),
			zap.String("color", color.String()),
			zap.Error(err))
		return
	}

	pub, exist := c.getPublisher(room.Code, color.Opposite())
	if !exist {
		c.logger.Warn("failed to find publisher for opponent, cannot notify of resignation",
			zap.String("room", room.Code),
			zap.String("winner", color.Opposite().String()),
			zap.String("resigner", color.String()),
		)
		return
	}

	err = pub.PublishJson(NewResignEvent(color))
	if err != nil {
		c.logger.Error("failed to publish resign event",
			zap.String("room", room.Code),
			zap.Error(err))
		return
	}

	room.signalGameOver()

	return
}

func (c *Coordinator) getPublisher(roomCode string, color engine.Color) (websocketPublisher, bool) {
	pubs, exist := c.roomPublishers[roomCode]
	if !exist {
		return nil, false
	}
	pub, exist := pubs[color]
	if !exist {
		return nil, false
	}
	return pub, true
}

type websocketPublisher interface {
	PublishJson(v any) error
}

type websocketConsumer interface {
	ConsumeJson(v any) error
}

type websocketContext interface {
	Context() context.Context
}

type websocketPublisherConsumer interface {
	websocketPublisher
	websocketConsumer
	websocketContext
}
