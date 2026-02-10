package room

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dyxj/chess/pkg/errorx"
	"github.com/dyxj/chess/pkg/httpx"
	"go.uber.org/zap"
)

type JoinHandler struct {
	logger     *zap.Logger
	repoFind   RepoFind
	repoUpdate RepoUpdate
}

type RepoFind interface {
	Find(code string) (*Room, bool)
}

type RepoUpdate interface {
	Update(room Room) error
}

type JoinRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (r *JoinRequest) Validate() *errorx.ValidationError {
	errs := make(map[string]string, 2)

	if len(r.Code) != 6 {
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

func NewJoinHandler(
	logger *zap.Logger,
	repoFind RepoFind,
	repoUpdate RepoUpdate,
) *JoinHandler {
	return &JoinHandler{
		logger:     logger,
		repoFind:   repoFind,
		repoUpdate: repoUpdate,
	}
}

func (h *JoinHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
}
