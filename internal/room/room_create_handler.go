package room

import (
	"net/http"
	"time"

	"github.com/dyxj/chess/pkg/httpx"
	"go.uber.org/zap"
)

type CreateHandler struct {
	logger  *zap.Logger
	creator Creator
}

type Creator interface {
	CreateRoom() (*Room, error)
}

type CreateResponse struct {
	Code        string    `json:"code"`
	Status      string    `json:"status"`
	CreatedTime time.Time `json:"createdTime"`
}

func NewCreateHandler(
	logger *zap.Logger,
	creator Creator,
) *CreateHandler {
	return &CreateHandler{
		logger:  logger,
		creator: creator,
	}
}

func (h *CreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	room, err := h.creator.CreateRoom()
	if err != nil {
		h.logger.Error("failed to create room", zap.Error(err))
		http.Error(w, "failed to create room", http.StatusInternalServerError)
		return
	}

	resp := CreateResponse{
		Code:        room.Code,
		Status:      room.Status.String(),
		CreatedTime: room.CreatedTime,
	}
	httpx.JsonResponse(http.StatusOK, resp, w)
}
