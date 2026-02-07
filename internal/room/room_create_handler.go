package room

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dyxj/chess/pkg/errorx"
	"github.com/dyxj/chess/pkg/httpx"
	"go.uber.org/zap"
)

const addRoomMaxRetries = 5

type CreateHandler struct {
	logger *zap.Logger
	repo   RepoAdd
}

type RepoAdd interface {
	Add(room Room) error
}

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

func NewCreateHandler(logger *zap.Logger, repo RepoAdd) *CreateHandler {
	return &CreateHandler{
		logger: logger,
		repo:   repo,
	}
}

func (h *CreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var cRequest CreateRequest
	err := json.NewDecoder(r.Body).Decode(&cRequest)
	if err != nil {
		h.logger.Warn("failed to decode request", zap.Error(err))
		httpx.BadRequestResponse("invalid request body",
			map[string]string{"error": err.Error()},
			w)
		return
	}

	vErr := cRequest.Validate()
	if vErr != nil {
		h.logger.Warn("validation failed", zap.Error(vErr))
		httpx.ValidationFailedResponse(vErr, w)
		return
	}

	_, err = h.createRoom(cRequest.PlayerName, color(cRequest.Color))
	if err != nil {
		h.logger.Error("failed to create room", zap.Error(err))
		httpx.InternalServerErrorResponse("", w)
		return
	}

	// TODO, upgrade to websocket
}

func (h *CreateHandler) createRoom(playerName string, color color) (Room, error) {

	room := NewEmptyRoom()
	room.SetPlayer(color, NewPlayer(playerName))

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
