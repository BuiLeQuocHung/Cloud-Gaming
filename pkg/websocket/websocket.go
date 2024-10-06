package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

type (
	Conn struct {
		mu sync.Mutex
		*websocket.Conn
	}
)

func New(conn *websocket.Conn) *Conn {
	return &Conn{
		Conn: conn,
		mu:   sync.Mutex{},
	}
}

func (c *Conn) WriteJSON(v interface{}) error {
	// handle "panic: concurrent write to websocket connection"
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.WriteJSON(v)
}
