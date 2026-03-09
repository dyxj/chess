package server

import (
	"net/http"

	"github.com/dyxj/chess/pkg/room"
	"go.uber.org/zap"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func BuildRouter(
	logger *zap.Logger,
	coordinator *room.Coordinator,
	corsEnabled bool,
) http.Handler {
	mux := http.NewServeMux()

	roomCreateHandler,
		roomJoinHandler,
		roomConnectHandler := setupRoomRoutes(logger, coordinator)

	mux.Handle("POST /room", roomCreateHandler)
	mux.Handle("POST /room/{code}/join", roomJoinHandler)
	mux.Handle("GET /room/connect", roomConnectHandler)

	if corsEnabled {
		return corsMiddleware(mux)
	}
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
