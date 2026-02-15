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
		"panic websocket manager Delete listener",
	)

	return m
}

func (m *Manager) Open(
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

	return &Publisher{key: key, conn: c}, &Consumer{key: key, conn: c, deleteChan: m.deleteChan}, nil
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

	err := conn.Close(statusCode, reason)
	if err != nil {
		m.logger.Warn("failed to close WebSocket", zap.String("key", key), zap.Error(err))
	}

	delete(m.conns, key)
}

func (m *Manager) CloseRead(ctx context.Context, key string) (context.Context, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, exist := m.getConn(key)
	if !exist {
		return nil, false
	}

	ctx = conn.CloseRead(ctx)

	return ctx, true
}

func (m *Manager) CloseNoHandshake(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger.Debug("closing websocket connection CloseNoHandshake", zap.String("player", key))

	conn, exist := m.getConn(key)
	if !exist {
		return
	}

	err := conn.CloseNow()
	if err != nil {
		m.logger.Warn("failed to close WebSocket", zap.String("key", key), zap.Error(err))
	}

	delete(m.conns, key)
}

func (m *Manager) isKeyExist(key string) bool {
	_, ok := m.conns[key]
	return ok
}

func (m *Manager) Delete(key string) {
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
		m.Delete(k)
	}
}

type Publisher struct {
	key  string
	conn *websocket.Conn
}

func (p *Publisher) Key() string {
	return p.key
}

func (p *Publisher) PublishJson(ctx context.Context, v any) error {
	return wsjson.Write(ctx, p.conn, v)
}

type Consumer struct {
	key        string
	conn       *websocket.Conn
	deleteChan chan<- string
}

func (c *Consumer) Key() string {
	return c.key
}

func (c *Consumer) ConsumeJson(ctx context.Context, v any) error {
	err := wsjson.Read(ctx, c.conn, v)
	if err != nil {
		// upon error, connection is closed with status websocket.StatusInvalidFramePayloadData
		c.deleteChan <- c.key
		return err
	}
	return nil
}
