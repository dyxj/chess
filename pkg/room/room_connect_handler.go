package room

import (
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

func (h *ConnectHandler) handleError(err error, w http.ResponseWriter) {
	h.logger.Error("failed to upgrade websocket connection", zap.Error(err))
	w.WriteHeader(http.StatusInternalServerError)
}
