package coordinator

import (
	"cloud_gaming/pkg/log"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func (c *Coordinator) workerRequestHandler(connection *Connection) {
	senderId := connection.id
	conn := connection.conn

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Debug("worker web socket closed", zap.Error(err))

			pair := c.binding.removeBinding(senderId)
			if pair == nil {
				break
			}

			workerConn := pair.worker.conn
			workerConn.Close()
			conn.Close()
			break
		}

		pair := c.binding.GetPair(senderId)
		if pair == nil {
			conn.Close()
			break
		}

		receiverConn := pair.user.conn
		receiverConn.WriteMessage(websocket.TextMessage, data)
	}
}
