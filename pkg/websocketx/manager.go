package websocketx

import (
	"errors"
	"net/http"
	"sync"

	"github.com/dyxj/chess/pkg/safe"
	"github.com/gobwas/ws"
	"go.uber.org/zap"
)

var (
	httpUpgrader ws.HTTPUpgrader
)

type Manager struct {
	logger      *zap.Logger
	mu          sync.RWMutex
	connections map[string]*Connection
	deleteChan  chan string
}

func NewManager(logger *zap.Logger) *Manager {
	m := &Manager{
		logger:      logger,
		connections: make(map[string]*Connection),
		deleteChan:  make(chan string, 100), // buffered channel to avoid blocking
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

	c, _, _, err := httpUpgrader.Upgrade(r, w)
	if err != nil {
		return nil, err
	}

	conn := NewConnection(key, c, m.deleteChan)
	m.addConn(key, conn)

	return conn, nil
}

func (m *Manager) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.connections, key)
}

// unexported helper methods

func (m *Manager) isKeyExist(key string) bool {
	_, ok := m.connections[key]
	return ok
}

func (m *Manager) addConn(key string, conn *Connection) {
	m.connections[key] = conn
}

func (m *Manager) deleteListener() {
	for k := range m.deleteChan {
		m.Delete(k)
	}
}
