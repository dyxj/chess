package server

import (
	"net/http"
	"time"

	"github.com/dyxj/chess/internal/room"
	"go.uber.org/zap"
)

func (s *Server) buildRouter() http.Handler {
	mux := http.NewServeMux()

	roomCreateHandler, roomJoinHandler := setupRoomRoutes(s.logger)

	mux.Handle("POST /room", roomCreateHandler)
	mux.Handle("POST /room/{code}/join", roomJoinHandler)
	// mux.Handle("GET /room/{id}/connect?ticket", ???)

	return mux
}

func setupRoomRoutes(
	logger *zap.Logger,
) (*room.CreateHandler, *room.JoinHandler) {
	coordinator := room.NewCoordinator(logger, 30*time.Second)
	creatorHandler := room.NewCreateHandler(logger, coordinator)
	joinHandler := room.NewJoinHandler(logger, coordinator)

	return creatorHandler, joinHandler
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
