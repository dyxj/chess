package room

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/dyxj/chess/pkg/engine"
	"github.com/dyxj/chess/pkg/errorx"
	"github.com/dyxj/chess/pkg/httpx"
	"go.uber.org/zap"
)

type JoinHandler struct {
	logger *zap.Logger
	joiner Joiner
}

type Joiner interface {
	IssueTicketToken(code string, name string, color engine.Color) (string, error)
}

type JoinRequest struct {
	Name  string       `json:"name"`
	Color engine.Color `json:"color"`
}

type JoinResponse struct {
	Token string `json:"token"`
}

const pathKeyCode = "code"

func NewJoinHandler(
	logger *zap.Logger,
	joiner Joiner,
) *JoinHandler {
	return &JoinHandler{
		logger: logger,
		joiner: joiner,
	}
}

func (h *JoinHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()
	code := r.PathValue(pathKeyCode)

	var joinReq JoinRequest
	err := json.NewDecoder(r.Body).Decode(&joinReq)
	if err != nil {
		h.logger.Warn("failed to decode join request", zap.Error(err))
		httpx.BadRequestResponse("invalid request body",
			map[string]string{"error": err.Error()},
			w)
		return
	}

	if err := h.validate(code, joinReq); err != nil {
		h.logger.Warn("invalid join room request", zap.Any("errors", err))
		httpx.ValidationFailedResponse(err, w)
		return
	}

	token, err := h.joiner.IssueTicketToken(code, joinReq.Name, joinReq.Color)
	if err != nil {
		h.handlerError(err, w)
		return
	}

	httpx.JsonResponse(http.StatusOK, JoinResponse{Token: token}, w)
}

func (h *JoinHandler) validate(code string, r JoinRequest) *errorx.ValidationError {
	errs := make(map[string]string, 2)

	if len(code) != 6 {
		errs["code"] = "code length must be 6 characters"
	}

	r.Name = strings.TrimSpace(r.Name)
	if r.Name == "" {
		errs["name"] = "name is required"
	}

	if len(errs) > 0 {
		return &errorx.ValidationError{Properties: errs}
	}

	return nil
}

func (h *JoinHandler) handlerError(err error, w http.ResponseWriter) {
	if errors.Is(err, ErrRoomNotFound) {
		httpx.NotFoundResponse(w)
		return
	}

	if errors.Is(err, ErrRoomFull) {
		httpx.BadRequestResponse("room is full", nil, w)
		return
	}

	h.logger.Error("failed to join room", zap.Error(err))
	httpx.InternalServerErrorResponse("", w)
}
