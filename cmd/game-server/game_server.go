package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/dyxj/chess/internal/adapter/server"
	"github.com/dyxj/chess/internal/config"
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
	router := server.BuildRouter(logger, memCache)
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
