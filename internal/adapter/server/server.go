package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/dyxj/chess/pkg/safe"
	"go.uber.org/zap"
)

type Server struct {
	logger *zap.Logger

	httpConfig HttpConfig

	httpServer *http.Server

	onGoingCtx            context.Context
	stopOngoingGracefully context.CancelFunc

	isShuttingDown atomic.Bool

	errSig  chan struct{}
	stopSig chan struct{}
	runDone chan struct{}
	done    chan struct{}
}

func NewServer(
	logger *zap.Logger,
	httpConfig HttpConfig,
) *Server {
	return &Server{
		logger:     logger,
		httpConfig: httpConfig,
		errSig:     make(chan struct{}),
		stopSig:    make(chan struct{}),
		runDone:    make(chan struct{}),
		done:       make(chan struct{}),
	}
}

func (s *Server) initServer(router http.Handler) {
	addr := fmt.Sprintf("%v:%v", s.httpConfig.Host(), s.httpConfig.Port())

	s.onGoingCtx, s.stopOngoingGracefully = context.WithCancel(context.Background())

	s.httpServer = &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: s.httpConfig.ReadHeaderTimeout(),
		ReadTimeout:       s.httpConfig.ReadTimeout(),
		IdleTimeout:       s.httpConfig.IdleTimeout(),
		Handler:           router,
		BaseContext: func(_ net.Listener) context.Context {
			return s.onGoingCtx
		},
	}
}

func (s *Server) Run(router http.Handler) <-chan struct{} {
	s.initServer(router)

	safe.GoWithLog(
		func() {
			s.logger.Info("starting httpServer", zap.String("address", s.httpServer.Addr))
			err := s.httpServer.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				s.logger.Error("httpServer failed to listen and serve", zap.Error(err))
				s.errSig <- struct{}{}
			}
			s.logger.Info("httpServer closed")
			close(s.runDone)
		},
		s.logger,
		"panic in httpServer",
	)

	safe.GoWithLog(
		s.listenForStopAndOrchestrateShutdown,
		s.logger,
		"panic in shutdown orchestration",
	)

	return s.errSig
}

func (s *Server) listenForStopAndOrchestrateShutdown() {
	<-s.stopSig

	s.logger.Info("received shutdown signal")

	// Set health check to unavailable, stops load balancers from sending new requests
	s.isShuttingDown.Store(true)

	// Allow time for readiness probe to pick up the change
	time.Sleep(s.httpConfig.ShutDownReadyDelay())

	// TODO should not operate based on timeout but rather when all timed games
	//  though some limitations needs to be added to game beforehand
	//  particular important for timed games
	//  moving games to storage for migration could work for
	//  games with no time control
	shutDownCtx, shutDownCancel := context.WithTimeout(context.Background(), s.httpConfig.ShutDownTimeout())
	defer shutDownCancel()

	// Stop receiving new requests and wait for ongoing requests to finish
	err := s.shutDown(shutDownCtx)
	s.stopOngoingGracefully()
	if err != nil {
		// In the event of force shutdown we do not wait for runDone.
		s.logger.Error("failed to wait for ongoing requests to finish, waiting for forced cancellation", zap.Error(err))
		time.Sleep(s.httpConfig.ShutDownHardTimeout())
		s.logger.Error("httpServer shut down ungracefully")
		close(s.done)
		return
	}

	s.logger.Info("httpServer shut down gracefully")
	<-s.runDone
	close(s.done)
}

func (s *Server) Stop() <-chan struct{} {
	close(s.stopSig)
	return s.done
}

func (s *Server) shutDown(shutDownCtx context.Context) error {
	s.logger.Info("initiate httpServer shutdown and wait for ongoing requests to finish")
	return s.httpServer.Shutdown(shutDownCtx)
}
