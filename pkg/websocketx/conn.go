package websocketx

import (
	"context"
	"encoding/json"
	"net"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Connection struct {
	key        string
	conn       net.Conn
	deleteChan chan<- string
	ctx        context.Context
	cancelFunc context.CancelFunc
	mu         sync.Mutex
	isClosed   bool
}

func NewConnection(
	key string,
	conn net.Conn,
	deleteChan chan<- string,
) *Connection {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &Connection{
		key:        key,
		conn:       conn,
		ctx:        ctx,
		cancelFunc: cancelFunc,
		deleteChan: deleteChan,
	}
}

func (c *Connection) Key() string {
	return c.key
}

func (c *Connection) PublishJson(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return wsutil.WriteServerText(c.conn, data)
}

func (c *Connection) ConsumeJson(v any) error {
	data, err := wsutil.ReadClientText(c.conn)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// WriteCloseStatusCode writes a close frame with the given status code and message
// The first close status written will be communicated to the client, subsequent calls will be ignored.
// Does not close the connection, caller should call Close() to close the connection after writing close status code.
func (c *Connection) WriteCloseStatusCode(code ws.StatusCode, msg string) error {
	err := wsutil.WriteServerMessage(c.conn, ws.OpClose, ws.NewCloseFrameBody(code, msg))
	if err != nil {
		return err
	}
	return nil
}

func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isClosed {
		return nil
	}
	c.isClosed = true
	err := c.conn.Close()
	c.notifyManagerToDeleteConn()
	return err
}

func (c *Connection) Cancel() {
	c.cancelFunc()
}

func (c *Connection) Context() context.Context {
	return c.ctx
}

func (c *Connection) notifyManagerToDeleteConn() {
	if c.deleteChan != nil {
		c.deleteChan <- c.key
	}
}
