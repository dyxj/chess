package server

import (
	"net/http"
	"time"

	"github.com/dyxj/chess/internal/room"
	"github.com/dyxj/chess/pkg/store"
	"go.uber.org/zap"
)

func BuildRouter(
	logger *zap.Logger,
	cache *store.MemCache,
) http.Handler {
	mux := http.NewServeMux()

	roomCreateHandler,
		roomJoinHandler,
		roomConnectHandler := setupRoomRoutes(logger, cache)

	mux.Handle("POST /room", roomCreateHandler)
	mux.Handle("POST /room/{code}/join", roomJoinHandler)
	mux.Handle("GET /room/connect", roomConnectHandler)

	return mux
}

func setupRoomRoutes(
	logger *zap.Logger,
	cache *store.MemCache,
) (*room.CreateHandler, *room.JoinHandler, *room.ConnectHandler) {
	coordinator := room.NewCoordinator(
		logger, 30*time.Second,
		room.NewMemCache(cache),
	)

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
