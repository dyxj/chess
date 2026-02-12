package server

import (
	"net/http"

	"github.com/dyxj/chess/internal/room"
	"go.uber.org/zap"
)

func (s *Server) buildRouter() http.Handler {
	mux := http.NewServeMux()

	roomCreateHandler := setupRoomRoutes(s.logger)

	mux.Handle("POST /room/create", roomCreateHandler)

	return mux
}

func setupRoomRoutes(
	logger *zap.Logger,
) *room.CreateHandler {
	coordinator := room.NewCoordinator(logger)
	creatorHandler := room.NewCreateHandler(logger, coordinator)

	return creatorHandler
}

/*
Handlers required
- New room
- Join room

Websocket operations
- Move
- Undo move
- Resign
- Force Draw
*/
