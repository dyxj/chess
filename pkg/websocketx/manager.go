package websocketx

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/dyxj/chess/pkg/safe"
	"go.uber.org/zap"
)

type Manager struct {
	logger     *zap.Logger
	mu         sync.RWMutex
	conns      map[string]*websocket.Conn
	deleteChan chan string
}

func NewManager(logger *zap.Logger) *Manager {
	m := &Manager{
		logger:     logger,
		conns:      make(map[string]*websocket.Conn),
		deleteChan: make(chan string, 100), // buffered channel to avoid blocking
	}

	safe.GoWithLog(
		m.deleteListener,
		logger,
		"panic websocket manager delete listener",
	)

	return m
}

func (m *Manager) OpenWebSocket(
	key string,
	w http.ResponseWriter,
	r *http.Request,
) (*Publisher, *Consumer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isKeyExist(key) {
		return nil, nil, errors.New("key already exists")
	}

	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return nil, nil, err
	}

	m.addConn(key, c)

	return &Publisher{conn: c}, &Consumer{key: key, conn: c, deleteChan: m.deleteChan}, nil
}

func (m *Manager) isKeyExist(key string) bool {
	_, ok := m.conns[key]
	return ok
}

func (m *Manager) deleteKey(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.conns, key)
}

func (m *Manager) addConn(key string, conn *websocket.Conn) {
	m.conns[key] = conn
}

func (m *Manager) getConn(key string) (*websocket.Conn, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, ok := m.conns[key]
	return conn, ok
}

func (m *Manager) deleteListener() {
	for k := range m.deleteChan {
		m.deleteKey(k)
	}
}

type Publisher struct {
	conn *websocket.Conn
}

func (p *Publisher) PublishJson(v any) error {
	return wsjson.Write(context.Background(), p.conn, v)
}

type Consumer struct {
	key        string
	conn       *websocket.Conn
	deleteChan chan<- string
}

func (c *Consumer) ConsumeJson(v any) error {
	err := wsjson.Read(context.Background(), c.conn, v)
	if err != nil {
		// upon error, connection is closed with status websocket.StatusInvalidFramePayloadData
		c.deleteChan <- c.key
		return err
	}
	return nil
}
