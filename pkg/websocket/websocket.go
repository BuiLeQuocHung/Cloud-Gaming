package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

type (
	Conn struct {
		mu sync.Mutex
		*websocket.Conn
		isConnected bool
	}
)

func New(conn *websocket.Conn) *Conn {
	return &Conn{
		Conn:        conn,
		mu:          sync.Mutex{},
		isConnected: true,
	}
}

func (c *Conn) WriteJSON(v interface{}) error {
	// handle "panic: concurrent write to websocket connection"
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.WriteJSON(v)
}

func (c *Conn) SetConnectionStatus(isConnected bool) {
	c.isConnected = isConnected
}

func (c *Conn) GetConnectionStatus() bool {
	return c.isConnected
}
