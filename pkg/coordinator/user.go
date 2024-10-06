package coordinator

import (
	"log"

	"github.com/gorilla/websocket"
)

func (c *Coordinator) userRequestHandler(connection *Connection) {
	senderId := connection.id
	conn := connection.conn

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Println("user web socket closed", err)
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
