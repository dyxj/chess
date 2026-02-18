package testx

import (
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"sync"

	"github.com/dyxj/chess/internal/adapter/server"
	"github.com/dyxj/chess/pkg/store"
	"go.uber.org/zap"
)

type Environment struct {
	name string

	ready       chan struct{}
	errorChan   chan error
	cleanupDone chan struct{}

	httptestServer *httptest.Server
	memCache       *store.MemCache

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

	httptestServer, err := e.buildHttpTestServer(e.memCache)
	if err != nil {
		errorChan <- err
		return
	}
	e.httptestServer = httptestServer

	close(ready)
}

func (e *Environment) buildHttpTestServer(memCache *store.MemCache) (*httptest.Server, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	router := server.BuildRouter(
		logger,
		memCache,
	)

	return httptest.NewServer(router), nil
}

func (e *Environment) closeHttpTestServer() {
	e.logger.Printf("close http test server")
	e.httptestServer.Close()
}

func (e *Environment) cleanup() {
	defer e.httptestServer.Close()
	e.closeHttpTestServer()
	close(e.cleanupDone)
}
