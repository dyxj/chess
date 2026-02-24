package server

import (
	"net/http"

	room2 "github.com/dyxj/chess/pkg/room"
	"go.uber.org/zap"
)

func BuildRouter(
	logger *zap.Logger,
	coordinator *room2.Coordinator,
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
	coordinator *room2.Coordinator,
) (*room2.CreateHandler, *room2.JoinHandler, *room2.ConnectHandler) {

	creatorHandler := room2.NewCreateHandler(logger, coordinator)
	joinHandler := room2.NewJoinHandler(logger, coordinator)
	connectHandler := room2.NewConnectHandler(logger, coordinator)

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
