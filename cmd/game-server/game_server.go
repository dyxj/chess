package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/dyxj/chess/internal/adapter/server"
	"github.com/dyxj/chess/internal/config"
	room2 "github.com/dyxj/chess/pkg/room"
	"github.com/dyxj/chess/pkg/store"
	"go.uber.org/zap"
)

func main() {
	// listen to interrupt and termination signals
	mainCtx, mainStop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// ensures stop function is called on exit to avoid unintended diversion of signals to context
	defer mainStop()

	logCfg := zap.NewProductionConfig()
	logCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	logger, err := logCfg.Build()
	if err != nil {
		log.Panicf("failed to initialize logger: %v", err)
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Panicf("failed to load config: %v", err)
	}

	httpServer := server.NewServer(logger, cfg.HTTPServerConfig)

	memCache := store.NewMemCache()

	coordinator := room2.NewCoordinator(
		logger, 30*time.Second,
		room2.NewMemCache(memCache),
	)
	router := server.BuildRouter(logger, coordinator)
	errSig := httpServer.Run(router)

	select {
	case <-mainCtx.Done():
	case <-errSig:
		logger.Error("unexpected error occurred while starting up httpServer")
	}
	mainStop()
	serverStopDone := httpServer.Stop()

	<-serverStopDone
}
