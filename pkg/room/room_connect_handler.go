package room

import (
	"errors"
	"net/http"

	"go.uber.org/zap"
)

type ConnectHandler struct {
	logger    *zap.Logger
	connector Connector
}

type Connector interface {
	ConnectWithToken(token string, w http.ResponseWriter, r *http.Request) error
}

func NewConnectHandler(
	logger *zap.Logger,
	connector Connector,
) *ConnectHandler {
	return &ConnectHandler{
		logger:    logger,
		connector: connector,
	}
}

const queryKeyToken = "token"

func (h *ConnectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	token := q.Get(queryKeyToken)

	err := h.connector.ConnectWithToken(token, w, r)
	if err != nil {
		h.handleError(err, w)
		return
	}
}

// websocket only returns status code pre-connection
func (h *ConnectHandler) handleError(err error, w http.ResponseWriter) {
	if errors.Is(err, ErrInvalidToken) {
		h.logger.Warn("invalid token", zap.Error(err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if errors.Is(err, ErrRoomNotFound) {
		h.logger.Warn("room not found", zap.Error(err))
		w.WriteHeader(http.StatusUnauthorized)
	}

	if errors.Is(err, ErrRoomFull) {
		h.logger.Warn("room full", zap.Error(err))
		w.WriteHeader(http.StatusConflict)
	}

	if errors.Is(err, ErrColorOccupied) {
		h.logger.Warn("color occupied", zap.Error(err))
		w.WriteHeader(http.StatusConflict)
	}

	h.logger.Error("failed to connect to room", zap.Error(err))
	w.WriteHeader(http.StatusInternalServerError)
}
