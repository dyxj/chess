package testx

import (
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"sync"
	"time"

	"github.com/dyxj/chess/internal/adapter/server"
	"github.com/dyxj/chess/internal/room"
	"github.com/dyxj/chess/pkg/store"
	"go.uber.org/zap"
)

type Environment struct {
	name string

	ready       chan struct{}
	errorChan   chan error
	cleanupDone chan struct{}

	httptestServer  *httptest.Server
	memCache        *store.MemCache
	roomCoordinator *room.Coordinator

	runOnce   sync.Once
	closeOnce sync.Once

	logger *log.Logger
}

func NewEnvironment(name string) *Environment {
	return &Environment{
		name:        name,
		ready:       make(chan struct{}),
		errorChan:   make(chan error),
		cleanupDone: make(chan struct{}),
		logger:      log.New(os.Stderr, fmt.Sprintf("test-env-%s ", name), log.LstdFlags),
	}
}

func (e *Environment) MemCache() *store.MemCache {
	return e.memCache
}

func (e *Environment) HTTTPTestServer() *httptest.Server {
	return e.httptestServer
}

func (e *Environment) RoomCoordinator() *room.Coordinator {
	return e.roomCoordinator
}

func (e *Environment) Run() (<-chan struct{}, <-chan error) {
	e.runOnce.Do(func() {
		e.logger.Printf("starting environment")
		go e.run(e.ready, e.errorChan)
	})

	return e.ready, e.errorChan
}

func (e *Environment) Close() {
	e.logger.Printf("closing environment")
	e.closeOnce.Do(func() {
		e.cleanup()
	})
	<-e.cleanupDone
	e.logger.Printf("environment closed")
}

func (e *Environment) run(ready chan struct{}, errorChan chan error) {
	e.memCache = store.NewMemCache()

	logger, err := zap.NewDevelopment()
	if err != nil {
		errorChan <- err
		return
	}

	e.roomCoordinator = room.NewCoordinator(
		logger,
		30*time.Second,
		room.NewMemCache(e.memCache),
	)

	httptestServer, err := e.buildHttpTestServer(e.roomCoordinator)
	if err != nil {
		errorChan <- err
		return
	}
	e.httptestServer = httptestServer

	close(ready)
}

func (e *Environment) buildHttpTestServer(roomCoordinator *room.Coordinator) (*httptest.Server, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	router := server.BuildRouter(
		logger,
		roomCoordinator,
	)

	return httptest.NewServer(router), nil
}

func (e *Environment) closeHttpTestServer() {
	if e.httptestServer == nil {
		return
	}
	e.logger.Printf("close http test server")
	e.httptestServer.Close()
}

func (e *Environment) cleanup() {
	e.closeHttpTestServer()
	close(e.cleanupDone)
}

func (e *Environment) Logger() *log.Logger {
	return e.logger
}
