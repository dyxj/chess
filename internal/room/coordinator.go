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
}

func NewCoordinator(
	logger *zap.Logger,
	tokenDuration time.Duration,
	cache *MemCache,
) *Coordinator {
	return &Coordinator{
		logger:        logger,
		cache:         cache,
		wsm:           websocketx.NewManager(logger),
		ticketCache:   NewTicketCache(),
		tokenDuration: tokenDuration,
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

	hasBoth := room.setPublisher(ticket.Color, conn)
	defer room.removePublisher(ticket.Color)

	logger := c.logger.With(
		zap.String("player", p.ID.String()),
		zap.String("room", room.Code),
		zap.String("color", ticket.Color.String()),
	)
	logger.Info("connected to room")

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

	if !hasBoth {
		err := c.publishEventMessage(conn, fmt.Sprintf("Waiting for %v player",
			ticket.Color.Opposite()))
		if err != nil {
			return err
		}
	} else {
		room.SetStatus(StatusInProgress)
		room.signalReady()
	}

	err = c.runLoop(room, ticket.Color, conn, logger)
	if err != nil {
		// errors after connection is established
		// should be communicated via socket
		c.handleRunLoopError(err, room, ticket.Color, conn, logger)
		return nil
	}

	return nil
}

func (c *Coordinator) runLoop(
	room *Room,
	color engine.Color,
	ws websocketPublisherConsumer,
	logger *zap.Logger,
) error {
	roundResultChan, consumeErrChan := c.goConsumeLoop(room, color, ws, logger)

	<-room.readyChan

	err := c.publishEventRound(ws, room.Game.Round())
	if err != nil {
		return err
	}

	processResultErrChan, err := c.goProcessRoundResults(room, color, roundResultChan, logger)
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
				continue
			}
			logger.Debug("exiting run loop, due to consume error")
			return consumeErr
		case processResultErr, ok := <-processResultErrChan:
			if !ok {
				processResultErrChan = nil
				continue
			}
			logger.Debug("exiting run loop, due to process result error")
			return processResultErr
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
	e := NewEventRound(round)
	err := p.PublishJson(e)
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

func (c *Coordinator) publishEventResign(p websocketPublisher, resigner engine.Color) error {
	e := NewResignEvent(resigner)
	err := p.PublishJson(e)
	if err != nil {
		return err
	}
	return nil
}

// goProcessRoundResults listens to roundResultChan.
// broadcast round results to both players concurrently in random order.
//
// If the round result indicates game over, it signals the room and exits.
//
// If an error occurred while publishing to the acting color, the error is published to
// the errorChan and the function exits.
//
// If an error occurred while publishing to the opponent color, the error is logged
// but not published and does not exit.
// Besides network errors during publishing both publishers should face the same error.
// - In the case, of network errors the routines will detect network errors even without
// being notified by this process.
//   - In case of other errors, this would result in the acting color exiting and resigning.
//     Could consider introducing game.StateError, does not seem required at the moment.
func (c *Coordinator) goProcessRoundResults(
	room *Room,
	color engine.Color,
	roundResultChan <-chan game.RoundResult,
	logger *zap.Logger,
) (<-chan error, error) {
	errorChan := make(chan error)

	go func() {
		defer safe.RecoverWithLog(logger, "goProcessRoundResults")()
		defer close(errorChan)

		for roundResult := range roundResultChan {
			pubs := room.publishers()
			wg := &sync.WaitGroup{}
			wg.Add(len(pubs))

			var errColor error

			for pColor, pub := range pubs {
				// use concurrent publish to provide fairness
				// though it doesn't matter as much for chess
				// ranging over map provides random order
				go func() {
					lg := c.logger.With(
						zap.String("room", room.Code),
						zap.String("color", pColor.String()),
					)
					defer safe.RecoverWithLog(lg, "goProcessRoundResults:broadcast")()
					defer wg.Done()

					err := c.publishEventRound(pub, roundResult)
					if err != nil {
						if pColor == color {
							errColor = fmt.Errorf("round result broadcast failed: %w", err)
						}

						return
					}
				}()
			}
			wg.Wait()

			if errColor != nil {
				errorChan <- errColor
				return
			}

			if roundResult.State.IsGameOver() {
				room.signalGameOver()
				return
			}
		}
	}()

	return errorChan, nil
}

func (c *Coordinator) goConsumeLoop(
	room *Room,
	color engine.Color,
	ws websocketPublisherConsumer,
	logger *zap.Logger,
) (<-chan game.RoundResult, <-chan error) {

	roundResultChan := make(chan game.RoundResult, 5)
	errChan := make(chan error)
	go func() {
		defer safe.RecoverWithLog(logger, "goConsumeLoop")()
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
					errChan <- fmt.Errorf("consume action failed: %w", err)
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

func (c *Coordinator) handleRunLoopError(
	err error,
	room *Room,
	color engine.Color,
	statusWriter websocketCloseStatusWriter,
	logger *zap.Logger,
) {
	c.resignAndNotifyOpponent(room, color)

	if websocketx.IsNetworkClosedError(err) {
		logger.Info("network closed", zap.Error(err))
		return
	}

	if wsErr, isErr := websocketx.IsWebSocketClosedError(err); isErr {
		if wsErr.Code == ws.StatusNormalClosure {
			logger.Info("websocket closed by client", zap.Error(err))
		} else {
			logger.Error("websocket closed by client", zap.Error(err))
		}
		return
	}

	if invalidPayloadErr, ok := errors.AsType[*websocketx.InvalidPayloadError](err); ok {
		logger.Info("invalid payload received from client", zap.Error(err))
		ipErr := statusWriter.WriteCloseStatusCode(
			ws.StatusInvalidFramePayloadData,
			invalidPayloadErr.Unwrap().Error(),
		)
		if ipErr != nil {
			logger.Error("failed to write invalid payload", zap.Error(ipErr))
		}
		return
	}

	logger.Error("unexpected error in run loop", zap.Error(err))
	csErr := statusWriter.WriteCloseStatusCode(ws.StatusInternalServerError, "internal error")
	if csErr != nil {
		logger.Error("failed to close status", zap.Error(csErr))
	}
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

	pub, exist := room.publisher(color.Opposite())
	if !exist {
		c.logger.Warn("failed to find publisher for opponent, cannot notify of resignation",
			zap.String("room", room.Code),
			zap.String("winner", color.Opposite().String()),
			zap.String("resigner", color.String()),
		)
		return
	}

	err = c.publishEventResign(pub, color)
	if err != nil {
		c.logger.Error("failed to publish resign event",
			zap.String("room", room.Code),
			zap.Error(err))
		return
	}

	room.signalGameOver()

	return
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

type websocketCloseStatusWriter interface {
	WriteCloseStatusCode(code ws.StatusCode, message string) error
}
