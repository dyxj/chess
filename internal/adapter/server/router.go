package server

import (
	"net/http"

	"github.com/dyxj/chess/internal/room"
	"go.uber.org/zap"
)

func (s *Server) buildRouter() http.Handler {
	mux := http.NewServeMux()

	roomCreateHandler := setupRoomRoutes(s.logger)

	mux.Handle("GET /room/create", roomCreateHandler)

	return mux
}

func setupRoomRoutes(
	logger *zap.Logger,
) *room.CreateHandler {
	repo := room.NewMemCache()
	webSocketManager := room.NewWebSocketManager(logger)
	creatorHandler := room.NewCreateHandler(logger, repo, webSocketManager)
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
