package room

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dyxj/chess/internal/engine"
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

	publisher, consumer, err := c.wsm.OpenWebSocket(p.ID.String(), w, r)
	if err != nil {
		room.RemovePlayer(ticket.Color)
		return err
	}
	// TODO defer should disconnect

	c.runLoop(room, ticket.Color, publisher, consumer)

	return nil
}

func (c *Coordinator) runLoop(
	room *Room,
	color engine.Color,
	pub *websocketx.Publisher,
	con *websocketx.Consumer,
) {
	c.registerPublisher(room.Code, color, pub)

	// todo setup constant consumer

	if room.HasBothPlayers() {
		room.signalReady()
	} else {
		c.publishWaitingForPlayer(pub, color.Opposite())
	}
	<-room.readyChan

	// register publisher

}

func (c *Coordinator) publishWaitingForPlayer(
	p *websocketx.Publisher,
	emptyColor engine.Color,
) {
	e := NewEventMessage(
		fmt.Sprintf("Waiting for %v player", emptyColor),
	)
	err := p.PublishJson(e)
	if err != nil {
		c.logger.Error("failed to publish WaitingForPlayer", zap.Error(err))
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

func (c *Coordinator) broadcast(roomCode string) {
	// todo implement
	// remember to ensure fairness of delivery, though it does not matter
}

func (c *Coordinator) consume() {
	// validate player turn
	// if not turn publish error message

	// if yes apply move
	// if fail valid publish error message
	// if accepted publish RoundResult
}
