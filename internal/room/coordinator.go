package room

import (
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
	roomPublishers map[string]map[engine.Color]*websocketx.Publisher
}

func NewCoordinator(logger *zap.Logger, tokenDuration time.Duration) *Coordinator {
	return &Coordinator{
		logger:         logger,
		cache:          NewMemCache(),
		wsm:            websocketx.NewManager(logger),
		ticketCache:    NewTicketCache(),
		tokenDuration:  tokenDuration,
		roomPublishers: make(map[string]map[engine.Color]*websocketx.Publisher),
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

	publisher, consumer, err := c.wsm.Open(p.ID.String(), w, r)
	if err != nil {
		room.RemovePlayer(ticket.Color)
		return err
	}
	defer c.wsm.CloseNoHandshake(p.ID.String())

	err = c.runLoop(room, ticket.Color, publisher, consumer)
	if err != nil {
		return err
	}

	c.wsm.Close(p.ID.String(), websocket.StatusNormalClosure, "game end")
	return nil
}

func (c *Coordinator) runLoop(
	room *Room,
	color engine.Color,
	pub *websocketx.Publisher,
	con *websocketx.Consumer,
) error {
	c.registerPublisher(room.Code, color, pub)

	stopChan := make(chan struct{})
	roundResultChan := c.consumeLoopInBackground(room, color, con, stopChan, pub)

	if room.HasBothPlayers() {
		room.SetStatus(StatusInProgress)
		room.signalReady()
	} else {
		xColor := color.Opposite()
		c.publishEventMessage(pub, fmt.Sprintf("Waiting for %v player", xColor))
	}
	<-room.readyChan

	c.publishEventRound(pub, room.Game.Round())

	c.setupRoundResultBroadcaster(room.Code, roundResultChan)

	return nil
}

// TODO error handling
//
//	Which errors should affect the process?
//	should check websocket.CloseStatus(err)
//	should check context, publish json currently doesn't provide context (context.Canceled)
func (c *Coordinator) publishEventMessage(p *websocketx.Publisher, msg string) {
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
func (c *Coordinator) publishEventRound(p *websocketx.Publisher, round game.RoundResult) {
	err := p.PublishJson(round)
	if err != nil {
		c.logger.Error("failed to publish RoundResult", zap.Error(err))
	}
}

// TODO error handling
func (c *Coordinator) publishEventError(p *websocketx.Publisher, eErr error) {
	e := NewEventError(eErr)
	err := p.PublishJson(e)
	if err != nil {
		c.logger.Error("failed to publish error event", zap.Any("eErr", eErr), zap.Error(err))
	}
}

func (c *Coordinator) registerPublisher(roomCode string, color engine.Color, pub *websocketx.Publisher) {
	c.muRoomPubs.Lock()
	defer c.muRoomPubs.Unlock()

	pubs, ok := c.roomPublishers[roomCode]
	if !ok {
		pubs = make(map[engine.Color]*websocketx.Publisher)
		c.roomPublishers[roomCode] = pubs
	}
	// overwriting existing is expected in case of reconnection
	pubs[color] = pub
}

func (c *Coordinator) setupRoundResultBroadcaster(roomCode string, roundResultChan <-chan game.RoundResult) {
	pubs, exist := c.roomPublishers[roomCode]
	if !exist {
		return
	}

	for roundResult := range roundResultChan {
		for color, pub := range pubs {
			// use concurrent publish to provide fairness
			// though it doesn't matter as much for chess
			// ranging over map provides random order
			safe.GoWithLog(
				func() {
					err := pub.PublishJson(roundResult)
					if err != nil {
						c.logger.Error("failed to broadcast",
							zap.String("room", roomCode),
							zap.String("player", pub.Key()),
							zap.String("color", color.String()),
							zap.Error(err))
					}
				},
				c.logger, "broadcast panic",
			)
		}
	}
}

// todo revisit error handling
func (c *Coordinator) broadcast(roomCode string, event any) {
	pubs, exist := c.roomPublishers[roomCode]
	if !exist {
		return
	}
	c.logger.Debug("broadcast", zap.Any("event", event))

	for color, pub := range pubs {
		// use concurrent publish to provide fairness
		// though it doesn't matter as much for chess
		// ranging over map provides random order
		safe.GoWithLog(
			func() {
				err := pub.PublishJson(event)
				if err != nil {
					c.logger.Error("failed to broadcast",
						zap.String("room", roomCode),
						zap.String("player", pub.Key()),
						zap.String("color", color.String()),
						zap.Error(err))
				}
			},
			c.logger, "broadcast panic",
		)
	}
}

// TODO error here needs to be propagated. In some cases the socket closes, it is crucial to notify parent
func (c *Coordinator) consumeLoopInBackground(
	room *Room,
	color engine.Color,
	con *websocketx.Consumer,
	stop <-chan struct{},
	pub *websocketx.Publisher,
) <-chan game.RoundResult {

	roundResultChan := make(chan game.RoundResult)

	safe.GoWithLog(
		func() {
			for {
				select {
				case <-stop:
					return
				default:
					var partial ActionPartial
					err := con.ConsumeJson(&partial)
					if err != nil {
						c.logger.Error("failed to consume", zap.Error(err))
						return
					}

					switch partial.Type {
					case ActionTypeMove:
						err = c.consumeAction(room, color, partial.Payload, pub, roundResultChan)
						if err != nil {
							return
						}
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
	pub *websocketx.Publisher,
	roundResultChan chan<- game.RoundResult,
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
		return c.processMoveAction(room, color, payload, pub, roundResultChan)
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
	pub *websocketx.Publisher,
	roundResultChan chan<- game.RoundResult,
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
