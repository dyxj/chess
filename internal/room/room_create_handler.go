package room

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dyxj/chess/internal/engine"
	"github.com/dyxj/chess/internal/game"
	"github.com/dyxj/chess/pkg/errorx"
	"github.com/dyxj/chess/pkg/httpx"
	"github.com/dyxj/chess/pkg/websocketx"
	"go.uber.org/zap"
)

const addRoomMaxRetries = 5

type CreateHandler struct {
	logger    *zap.Logger
	repo      CreatorRepo
	wsCreator WebsocketCreator
}

type CreatorRepo interface {
	Add(room Room) error
	Delete(room Room)
}

type WebsocketCreator interface {
	OpenWebSocket(
		key string, w http.ResponseWriter, r *http.Request,
	) (*websocketx.Publisher, *websocketx.Consumer, error)
}

const queryKeyPlayerName = "playerName"
const queryKeyColor = "color"

type CreateRequest struct {
	PlayerName string `json:"playerName"`
	Color      string `json:"color"`
}

type CreateResponse struct {
	Code        string    `json:"code"`
	Status      string    `json:"status"`
	CreatedTime time.Time `json:"createdTime"`
}

func (r *CreateRequest) Validate() *errorx.ValidationError {
	errs := make(map[string]string, 2)

	r.PlayerName = strings.TrimSpace(r.PlayerName)
	if r.PlayerName == "" {
		errs["playerName"] = "name is required"
	}

	if r.Color != white.String() && r.Color != black.String() {
		errs["color"] = "color must be either 'white' or 'black'"
	}

	if len(errs) > 0 {
		return &errorx.ValidationError{Properties: errs}
	}

	return nil
}

func NewCreateHandler(
	logger *zap.Logger,
	repo CreatorRepo,
	wsCreator WebsocketCreator,
) *CreateHandler {
	return &CreateHandler{
		logger:    logger,
		repo:      repo,
		wsCreator: wsCreator,
	}
}

func (h *CreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	cRequest := &CreateRequest{
		PlayerName: q.Get(queryKeyPlayerName),
		Color:      q.Get(queryKeyColor),
	}

	vErr := cRequest.Validate()
	if vErr != nil {
		h.logger.Warn("validation failed", zap.Error(vErr))
		httpx.ValidationFailedResponse(vErr, w)
		return
	}

	rColor := color(cRequest.Color)
	room, err := h.createRoom(cRequest.PlayerName, rColor)
	if err != nil {
		h.logger.Error("failed to create room", zap.Error(err))
		httpx.InternalServerErrorResponse("", w)
		return
	}

	rKey := room.connectionKey(rColor)
	publisher, consumer, err := h.wsCreator.OpenWebSocket(rKey, w, r)
	if err != nil {
		h.logger.Error("failed to establish websocket connection", zap.Error(err))
		h.repo.Delete(room)
		httpx.InternalServerErrorResponse("", w)
		return
	}

	// everything below here should be moved to a coordinator
	err = publisher.PublishJson(Event{
		Status:    StatusWaiting,
		Message:   "Waiting for player to join",
		GameState: game.StateInProgress,
		Move: game.Move{
			Color:  engine.White,
			Symbol: engine.Pawn,
			From:   0,
			To:     2,
		},
	})
	if err != nil {
		h.logger.Error("failed to publish event", zap.Error(err))
	}

	var cErr error
	for cErr == nil {
		var action Action
		cErr = consumer.ConsumeJson(&action)
		if cErr != nil {
			continue
		}
		h.logger.Info("websocket connection published", zap.Any("action", action))
	}
}

func (h *CreateHandler) createRoom(playerName string, color color) (Room, error) {

	room := NewEmptyRoom()
	room.setPlayer(color, NewPlayer(playerName))

	retry := 0
	for {
		err := h.repo.Add(*room)
		if err == nil {
			return *room, nil
		}

		if !errors.Is(err, ErrCodeAlreadyExists) {
			return Room{}, fmt.Errorf("failed to create room due to unexpected error: %w", err)
		}

		if retry >= addRoomMaxRetries {
			return Room{}, fmt.Errorf("failed to create room after %d attempts: %w", retry, err)
		}

		h.logger.Warn("failed to add room to repo, retrying", zap.Error(err), zap.Int("retry", retry))
		retry++
		room.Code = generateCode()
	}
}
