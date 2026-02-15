package websocketx

import (
	"errors"
	"net/http"
	"sync"

	"github.com/coder/websocket"
	"github.com/dyxj/chess/pkg/safe"
	"go.uber.org/zap"
)

type Manager struct {
	logger     *zap.Logger
	mu         sync.RWMutex
	conns      map[string]*Connection
	deleteChan chan string
}

func NewManager(logger *zap.Logger) *Manager {
	m := &Manager{
		logger:     logger,
		conns:      make(map[string]*Connection),
		deleteChan: make(chan string, 100), // buffered channel to avoid blocking
	}

	safe.GoWithLog(
		m.deleteListener,
		logger,
		"panic websocket manager Delete listener",
	)

	return m
}

func (m *Manager) Open(
	key string,
	w http.ResponseWriter,
	r *http.Request,
) (*Connection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isKeyExist(key) {
		return nil, errors.New("key already exists")
	}

	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return nil, err
	}

	conn := NewConnection(key, c, m.deleteChan)
	m.addConn(key, conn)

	return conn, nil
}

func (m *Manager) Close(
	key string,
	statusCode websocket.StatusCode,
	reason string,
) {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, exist := m.getConn(key)
	if !exist {
		return
	}

	err := conn.conn.Close(statusCode, reason)
	if err != nil {
		m.logger.Warn("failed to close WebSocket", zap.String("key", key), zap.Error(err))
	}

	delete(m.conns, key)
}

func (m *Manager) CloseNoHandshake(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger.Debug("closing websocket connection CloseNoHandshake", zap.String("player", key))

	conn, exist := m.getConn(key)
	if !exist {
		return
	}

	err := conn.conn.CloseNow()
	if err != nil {
		m.logger.Warn("failed to close WebSocket", zap.String("key", key), zap.Error(err))
	}

	delete(m.conns, key)
}

func (m *Manager) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.conns, key)
}

// unexported helper methods

func (m *Manager) isKeyExist(key string) bool {
	_, ok := m.conns[key]
	return ok
}

func (m *Manager) addConn(key string, conn *Connection) {
	m.conns[key] = conn
}

func (m *Manager) getConn(key string) (*Connection, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, ok := m.conns[key]
	return conn, ok
}

func (m *Manager) deleteListener() {
	for k := range m.deleteChan {
		m.Delete(k)
	}
}
