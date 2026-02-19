package server

import (
	"net/http"

	"github.com/dyxj/chess/internal/room"
	"go.uber.org/zap"
)

func BuildRouter(
	logger *zap.Logger,
	coordinator *room.Coordinator,
) http.Handler {
	mux := http.NewServeMux()

	roomCreateHandler,
		roomJoinHandler,
		roomConnectHandler := setupRoomRoutes(logger, coordinator)

	mux.Handle("POST /room", roomCreateHandler)
	mux.Handle("POST /room/{code}/join", roomJoinHandler)
	mux.Handle("GET /room/connect", roomConnectHandler)

	return mux
}

func setupRoomRoutes(
	logger *zap.Logger,
	coordinator *room.Coordinator,
) (*room.CreateHandler, *room.JoinHandler, *room.ConnectHandler) {

	creatorHandler := room.NewCreateHandler(logger, coordinator)
	joinHandler := room.NewJoinHandler(logger, coordinator)
	connectHandler := room.NewConnectHandler(logger, coordinator)

	return creatorHandler, joinHandler, connectHandler
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
