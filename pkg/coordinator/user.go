package coordinator

import (
	"cloud_gaming/pkg/log"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func (c *Coordinator) userRequestHandler(connection *Connection) {
	senderId := connection.id
	conn := connection.conn

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Debug("user web socket closed", zap.Error(err))
			conn.SetConnectionStatus(false)

			pair := c.binding.RemoveBinding(senderId)
			if pair == nil {
				break
			}

			c.freeWorkers <- pair.worker
			conn.Close()
			break
		}

		pair := c.binding.GetPair(senderId)
		if pair == nil {
			conn.Close()
			break
		}

		receiverConn := pair.worker.conn
		receiverConn.WriteMessage(websocket.TextMessage, data)
	}
}
